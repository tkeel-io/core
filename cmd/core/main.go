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
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	corev1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/dispatch"
	"github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/kafka"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/loopback"
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/noop"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
	_ "github.com/tkeel-io/core/pkg/resource/store/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/store/noop"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/influxdb"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/noop"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/util/discovery"
	_ "github.com/tkeel-io/core/pkg/util/transport"
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

var _apiManager apim.APIManager
var _dispatcher dispatch.Dispatcher

func main() {
	cmd := cobra.Command{
		Use:     "core",
		Short:   "Start a new core runtime",
		Example: _coreCmdExample,
		Run:     core,
	}

	cmd.PersistentFlags().StringVarP(&_cfgFile, "conf", "c", "config.yml", "config file path.")
	cmd.PersistentFlags().StringVar(&_httpAddr, "http_addr", ":6789", "http listen address.")
	cmd.PersistentFlags().StringVar(&_grpcAddr, "grpc_addr", ":31234", "grpc listen address.")
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

	{
		// set default configurations.
		config.SetDefaultEtcd(_etcdBrokers)
		config.Init(_cfgFile)
	}

	// init gllbal placement.
	placement.Initialize()

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

	// new proxy.
	httpProxySrv := http.NewServer(fmt.Sprintf(":%d", config.Get().Proxy.HTTPPort))
	grpcProxySrv := grpc.NewServer(fmt.Sprintf(":%d", config.Get().Proxy.GRPCPort))
	serverList = append(serverList, httpProxySrv, grpcProxySrv)

	coreApp := app.New(config.Get().Server.AppID,
		&log.Conf{
			App:    config.Get().Server.AppID,
			Level:  config.Get().Logger.Level,
			Dev:    config.Get().Logger.Dev,
			Output: config.Get().Logger.Output,
		},
		serverList...,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceRegisterToCoreV1(ctx, httpSrv, grpcSrv)
	serviceRegisterToProxyV1(ctx, httpProxySrv, grpcProxySrv)

	if err := coreApp.Run(context.TODO()); err != nil {
		log.Fatal(err)
	}

	// wait sidecar ready.
	time.Sleep(1 * time.Second)

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

	// register service.
	if err = discoveryEnd.Register(
		context.Background(),
		discovery.Service{
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

	// create message dispatcher.
	coreRepo := repository.New(coreDao)
	if err = loadDispatcher(context.Background(), coreRepo); nil != err {
		log.Fatal(err)
	}

	var stateManager types.Manager
	if stateManager, err = runtime.NewManager(context.Background(), newResourceManager(coreRepo), _dispatcher); nil != err {
		log.Fatal(err)
	}

	if _apiManager, err = apim.New(context.Background(), coreRepo, _dispatcher); nil != err {
		log.Fatal(err)
	}

	if err = _apiManager.Start(); nil != err {
		log.Fatal(err)
	} else if err = stateManager.Start(); nil != err {
		log.Fatal(err)
	}

	// initialize core services.
	initialzeService(_apiManager, search.GlobalService)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err = coreApp.Stop(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func initialzeService(apiManager apim.APIManager, searchClient corev1.SearchHTTPServer) {
	// initialize entity service.
	_entitySrv.Init(apiManager, searchClient)
	// initialize subscription service.
	_subscriptionSrv.Init(apiManager)
	// initialize topic service.
	_topicSrv.Init(apiManager)
	// initialize search service.
	_searchSrv.Init(searchClient)
	// initialize proxy service.
	_proxySrv.Init(apiManager)
}

var (
	_tsSrv           *service.TSService
	_topicSrv        *service.TopicService
	_proxySrv        *service.ProxyService
	_entitySrv       *service.EntityService
	_searchSrv       *service.SearchService
	_subscriptionSrv *service.SubscriptionService
)

// serviceRegisterToCoreV1 register your services here.
func serviceRegisterToCoreV1(ctx context.Context, httpSrv *http.Server, grpcSrv *grpc.Server) {
	var err error
	// register entity service.
	if _entitySrv, err = service.NewEntityService(ctx); nil != err {
		log.Fatal(err)
	}
	corev1.RegisterEntityHTTPServer(httpSrv.Container, _entitySrv)
	corev1.RegisterEntityServer(grpcSrv.GetServe(), _entitySrv)

	// register subscription service.
	if _subscriptionSrv, err = service.NewSubscriptionService(ctx); nil != err {
		log.Fatal(err)
	}
	corev1.RegisterSubscriptionHTTPServer(httpSrv.Container, _subscriptionSrv)
	corev1.RegisterSubscriptionServer(grpcSrv.GetServe(), _subscriptionSrv)

	// register topic service.
	if _topicSrv, err = service.NewTopicService(ctx); nil != err {
		log.Fatal(err)
	}
	corev1.RegisterTopicHTTPServer(httpSrv.Container, _topicSrv)
	corev1.RegisterTopicServer(grpcSrv.GetServe(), _topicSrv)

	// register search service.
	_searchSrv = service.NewSearchService()
	corev1.RegisterSearchHTTPServer(httpSrv.Container, _searchSrv)
	corev1.RegisterSearchServer(grpcSrv.GetServe(), _searchSrv)

	// register search service.
	_tsSrv = service.NewTSService()
	corev1.RegisterTSHTTPServer(httpSrv.Container, _tsSrv)
}

func serviceRegisterToProxyV1(ctx context.Context, httpSrv *http.Server, grpcSrv *grpc.Server) {
	// register proxy service.
	_proxySrv = service.NewProxyService()
	corev1.RegisterProxyHTTPServer(httpSrv.Container, _proxySrv)
	corev1.RegisterProxyServer(grpcSrv.GetServe(), _proxySrv)
}

func newResourceManager(coreRepo repository.IRepository) types.ResourceManager {
	log.Info("create core default resources")
	// default time series.
	tsdbClient := tseries.NewTimeSerier(resource.ParseFrom(config.Get().Components.TimeSeries).Name)

	return runtime.NewResources(search.GlobalService, tsdbClient, coreRepo)
}

func loadDispatcher(ctx context.Context, repo repository.IRepository) error {
	log.Info("load local Queues.")
	// load loacal Queues.
	for _, queue := range config.Get().Dispatcher.Queues {
		properties := make(map[string]interface{})
		for _, pair := range queue.Metadata {
			properties[pair.Key] = pair.Value
		}

		if queue.Version > 0 {
			// compare version.
			var err error
			var remoteQueue *dao.Queue
			if remoteQueue, err = repo.GetQueue(context.Background(), &dao.Queue{ID: queue.ID}); nil != err {
				log.Warn("query Queue", zap.Error(err), logger.ID(queue.ID))
				continue
			}

			if remoteQueue.Version >= queue.Version {
				continue
			}
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

	dispatcher := dispatch.New(
		context.Background(),
		config.Get().Dispatcher.ID,
		config.Get().Dispatcher.Name,
		config.Get().Dispatcher.Enabled, repo)

	if err := dispatcher.Run(); nil != err {
		log.Error("run dispatcher", zap.Error(err), logger.ID(config.Get().Dispatcher.ID))
		return errors.Wrap(err, "start dispatcher")
	}

	_dispatcher = dispatcher

	return nil
}
