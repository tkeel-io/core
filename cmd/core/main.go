/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	corev1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
	_ "github.com/tkeel-io/core/pkg/resource/state/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/state/noop"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/influxdb"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/noop"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/server"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/version"

	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"
	"github.com/tkeel-io/kit/transport/grpc"
	"github.com/tkeel-io/kit/transport/http"
)

const _coreCmdExample = `you can use this like following:
set your config file
# default 'config.yml'
core -c <config file> 
core --conf <config file> 

core --http_addr <http address and port>
core --grpc_addr <grpc address and port>

configure your etcd server, you can specify multiple
core --etcd <one etcd server address>

configure your elasticsearch server, you can specify multiple
core --search-engine drive://username:password@url0,url1
`

var (
	_cfgFile      string
	_httpAddr     string
	_grpcAddr     string
	_etcdBrokers  []string
	_searchEngine string
)

var _entityManager entities.EntityManager

func main() {
	cmd := cobra.Command{
		Use:     "core",
		Short:   "Start a new core runtime",
		Example: _coreCmdExample,
		Run:     core,
	}

	cmd.PersistentFlags().StringVarP(&_cfgFile, "conf", "c", "config.yml", "config file path.")
	cmd.PersistentFlags().StringVar(&_httpAddr, "http_addr", ":6789", "http listen address.")
	cmd.PersistentFlags().StringVar(&_grpcAddr, "grpc_addr", ":31233", "grpc listen address.")
	cmd.PersistentFlags().StringSliceVar(&_etcdBrokers, "etcd", nil, "etcd brokers address.")
	cmd.PersistentFlags().StringVar(&_searchEngine, "search-engine", "", "your search engine SDN.")
	cmd.Version = version.Version
	cmd.SetVersionTemplate(version.Template())

	{
		// Subcommand register here.
		cmd.AddCommand()
	}

	cobra.OnInitialize(func() {
		// Some initialize Func
		// called after flags have been initialized
		// before the subcommand execute.
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func core(cmd *cobra.Command, args []string) {
	logger.InfoStatusEvent(os.Stdout, "loading configuration...")
	config.Init(_cfgFile)

	// user flags input recover config file content.
	if _etcdBrokers != nil {
		config.SetEtcdBrokers(_etcdBrokers)
	}

	// rewrite search engine config by flags input info.
	if _searchEngine != "" {
		drive, username, password, urls, err := util.ParseSearchEngine(_searchEngine)
		if err != nil {
			logger.FailureStatusEvent(os.Stdout, "please check your --search-engine configuration(driver://username:password@url1,url2)")
			return
		}
		switch drive {
		case driver.ElasticsearchDriver:
			config.SetSearchEngineElasticsearchConfig(username, password, urls)
			// add use flag when more drive flags are available.
			config.SetSearchEngineUseDrive(string(drive))
		}
	}

	// Start Search Service.
	search.GlobalService = search.Init()
	// logger started Info.
	switch config.Get().SearchEngine.Use {
	case string(driver.ElasticsearchDriver):
		logger.InfoStatusEvent(os.Stdout, "Success init Elasticsearch Service for Search Engine")
	}

	// new servers.
	httpSrv := server.NewHTTPServer(_httpAddr)
	grpcSrv := server.NewGRPCServer(_grpcAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	coreApp := app.New(config.Get().Server.AppID,
		&log.Conf{
			App:    config.Get().Server.AppID,
			Level:  config.Get().Logger.Level,
			Dev:    config.Get().Logger.Dev,
			Output: config.Get().Logger.Output,
		},
		serverList...,
	)

	var err error
	var coreDao *dao.Dao
	var coreRepo repository.IRepository
	coreDao, err = dao.New(context.Background(), config.Get().Store, config.Get().Etcd)
	if nil != err {
		log.Fatal(err)
	}

	coreRepo = repository.New(coreDao)
	stateManager, err := runtime.NewManager(context.Background(), coreRepo)
	if nil != err {
		log.Fatal(err)
	}

	_entityManager, err = entities.NewEntityManager(context.Background(), coreRepo, stateManager, search.GlobalService)
	if nil != err {
		log.Fatal(err)
	}

	serviceRegisterToCoreV1(httpSrv, grpcSrv)

	logger.SuccessStatusEvent(os.Stdout, "all service registered.")
	logger.SuccessStatusEvent(os.Stdout, "everything is ready for execution.")
	if err = _entityManager.Start(); nil != err {
		log.Fatal(err)
	}
	if err = coreApp.Run(context.TODO()); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err = coreApp.Stop(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

// serviceRegisterToCoreV1 register your services here.
func serviceRegisterToCoreV1(httpSrv *http.Server, grpcSrv *grpc.Server) {
	// register entity service.
	EntitySrv, err := service.NewEntityService(context.Background(), _entityManager, search.GlobalService)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)
	corev1.RegisterEntityServer(grpcSrv.GetServe(), EntitySrv)

	// register subscription service.
	SubscriptionSrv, err := service.NewSubscriptionService(context.Background(), _entityManager)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterSubscriptionHTTPServer(httpSrv.Container, SubscriptionSrv)
	corev1.RegisterSubscriptionServer(grpcSrv.GetServe(), SubscriptionSrv)

	// register topic service.
	TopicSrv, err := service.NewTopicService(context.Background(), _entityManager)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterTopicHTTPServer(httpSrv.Container, TopicSrv)
	corev1.RegisterTopicServer(grpcSrv.GetServe(), TopicSrv)

	// register search service.
	SearchSrv := service.NewSearchService(search.GlobalService)
	corev1.RegisterSearchHTTPServer(httpSrv.Container, SearchSrv)
	corev1.RegisterSearchServer(grpcSrv.GetServe(), SearchSrv)
}
