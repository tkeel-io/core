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
	"github.com/tkeel-io/core/pkg/dispatch"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/noop"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
	_ "github.com/tkeel-io/core/pkg/resource/state/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/state/noop"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/influxdb"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/noop"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/util/discovery"
	"github.com/tkeel-io/core/pkg/version"
	"go.uber.org/zap"

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
		log.Fatalf("execute core service failed, %s", err.Error())
	}
}

func core(cmd *cobra.Command, args []string) {
	logger.InfoStatusEvent(os.Stdout, "loading configuration...")
	config.Init(_cfgFile)

	// init gllbal placement.
	placement.Initialize()

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
	switch config.Get().Components.SearchEngine.Use {
	case string(driver.ElasticsearchDriver):
		logger.InfoStatusEvent(os.Stdout, "Success init Elasticsearch Service for Search Engine")
	}

	// new servers.
	httpSrv := http.NewServer(_httpAddr)
	grpcSrv := grpc.NewServer(_grpcAddr)
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

	// register core service.
	var err error
	var discoveryEnd *discovery.Discovery
	if discoveryEnd, err = discovery.New(discovery.Config{
		Endpoints:   config.Get().Discovery.Endpoints,
		HeartTime:   config.Get().Discovery.HeartTime,
		DialTimeout: config.Get().Discovery.DialTimeout,
	}); nil != err {
		log.Fatal(err)
	}

	// core run context.Context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// register service.
	if err = discoveryEnd.Register(ctx, discovery.Service{
		Name:     config.Get().Server.Name,
		AppID:    config.Get().Server.AppID,
		Port:     config.Get().Server.AppPort,
		Host:     util.ResolveAddr(),
		Metadata: map[string]string{},
	}); nil != err {
		log.Fatal(err)
	}

	var coreDao *dao.Dao
	if coreDao, err = dao.New(ctx, config.Get().Components.Store, config.Get().Components.Etcd); nil != err {
		log.Fatal(err)
	}

	coreRepo := repository.New(coreDao)
	var stateManager statem.StateManager
	if stateManager, err = runtime.NewManager(context.Background(), newResourceManager(coreRepo)); nil != err {
		log.Fatal(err)
	}

	if _entityManager, err = entities.NewEntityManager(context.Background(), coreRepo, stateManager); nil != err {
		log.Fatal(err)
	}

	// create message dispatcher.
	loadDispatcher(context.Background(), coreRepo)

	serviceRegisterToCoreV1(ctx, httpSrv, grpcSrv)

	if err = _entityManager.Start(); nil != err {
		log.Fatal(err)
	} else if err = coreApp.Run(context.TODO()); err != nil {
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
func serviceRegisterToCoreV1(ctx context.Context, httpSrv *http.Server, grpcSrv *grpc.Server) {
	// register entity service.
	EntitySrv, err := service.NewEntityService(ctx, _entityManager, search.GlobalService)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)
	corev1.RegisterEntityServer(grpcSrv.GetServe(), EntitySrv)

	// register subscription service.
	SubscriptionSrv, err := service.NewSubscriptionService(ctx, _entityManager)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterSubscriptionHTTPServer(httpSrv.Container, SubscriptionSrv)
	corev1.RegisterSubscriptionServer(grpcSrv.GetServe(), SubscriptionSrv)

	// register topic service.
	TopicSrv, err := service.NewTopicService(ctx, _entityManager)
	if nil != err {
		log.Fatal(err)
	}
	corev1.RegisterTopicHTTPServer(httpSrv.Container, TopicSrv)
	corev1.RegisterTopicServer(grpcSrv.GetServe(), TopicSrv)

	// register search service.
	SearchSrv := service.NewSearchService(search.GlobalService)
	corev1.RegisterSearchHTTPServer(httpSrv.Container, SearchSrv)
	corev1.RegisterSearchServer(grpcSrv.GetServe(), SearchSrv)

	// register proxy service.
	ProxyService := service.NewProxyService(_entityManager)
	corev1.RegisterProxyHTTPServer(httpSrv.Container, ProxyService)
	corev1.RegisterProxyServer(grpcSrv.GetServe(), ProxyService)
}

func newResourceManager(coreRepo repository.IRepository) statem.ResourceManager {
	log.Info("create core default resources")
	// default pubsub.
	pubsubClient := pubsub.NewPubsub(resource.ParseFrom(config.Get().Components.Pubsub))
	// default time series.
	tsdbClient := tseries.NewTimeSerier(resource.ParseFrom(config.Get().Components.TimeSeries))

	return runtime.NewResources(pubsubClient, search.GlobalService, tsdbClient, coreRepo)
}

func loadDispatcher(ctx context.Context, repo repository.IRepository) {
	log.Info("load local Queues.")
	// load loacal Queues.
	for _, queue := range config.Get().Dispatcher.Queues {
		properties := make(map[string]interface{})
		for _, pair := range queue.Metadata {
			properties[pair.Key] = pair.Value
		}

		repo.PutQueue(ctx, &dao.Queue{
			ID:           queue.ID,
			Name:         queue.Name,
			Type:         dao.QueueType(queue.Type),
			Version:      queue.Version,
			NodeName:     config.Get().Server.Name,
			Consumers:    queue.Consumers,
			ConsumerType: dao.ConsumerType(queue.ConsumerType),
			Description:  queue.Description,
			Metadata:     properties,
		})
		log.Info("load local Queue", logger.ID(queue.ID),
			zap.String("consumer_type", queue.ConsumerType),
			logger.Type(queue.Type), logger.Version(queue.Version),
			logger.Name(queue.Name), logger.Desc(queue.Description),
			zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))
	}

	dispatcher := dispatch.NewDispatcher(context.Background(),
		config.Get().Dispatcher.ID, config.Get().Dispatcher.Name, repo)

	if err := dispatcher.Run(); nil != err {
		log.Fatal(err)
	}
}
