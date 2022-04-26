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
	"strconv"
	"strings"
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
	_ "github.com/tkeel-io/core/pkg/resource/pubsub/noop"
	"github.com/tkeel-io/core/pkg/resource/search"
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
	_cfgFile    string
	_apiManager apim.APIManager
	_dispatcher dispatch.Dispatcher
)

func main() {
	cmd := cobra.Command{
		Use:     "core",
		Short:   "Start a new core runtime",
		Example: _coreCmdExample,
		Run:     core,
	}

	cmd.Version = version.Version
	cmd.SetVersionTemplate(version.Template())
	cmd.PersistentFlags().StringVarP(&_cfgFile, "conf", "c", "config.yml", "config file path.")
	cmd.Flags().String("http_addr", ":6789", "core http server listen address.")
	cmd.Flags().String("grpc_addr", ":31234", "core http server listen address.")
	cmd.Flags().Int("proxy_http_port", 20000, "core proxy http listen address port.")
	cmd.Flags().Int("proxy_grpc_port", 20001, "core proxy http listen address port.")
	cmd.Flags().StringSlice("etcd", nil, "etcd brokers address, example: --etcd=\"http://localhost:2379,http://192.168.12.90:2379\"")
	cmd.Flags().String("search_engine", "", "your search engine SDN.")

	// bind commandline arguments.
	cmdViper := config.GetCmdV()
	cmdViper.BindPFlag("components.etcd.endpoints", cmd.Flags().Lookup("etcd"))
	cmdViper.BindPFlag("components.search_engine", cmd.Flags().Lookup("search_engine"))
	cmdViper.BindPFlag("server.http_addr", cmd.Flags().Lookup("http_addr"))
	cmdViper.BindPFlag("server.grpc_addr", cmd.Flags().Lookup("grpc_addr"))
	cmdViper.BindPFlag("proxy.http_port", cmd.Flags().Lookup("proxy_http_port"))
	cmdViper.BindPFlag("proxy.grpc_port", cmd.Flags().Lookup("proxy_grpc_port"))

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
	logger.InfoStatusEvent(os.Stdout, "configuration loaded")

	// init gllbal placement.
	placement.Initialize()

	// new servers.
	httpSrv := http.NewServer(config.Get().Server.HTTPAddr)
	grpcSrv := grpc.NewServer(config.Get().Server.GRPCAddr)
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
			Name:  config.Get().Server.Name,
			AppID: config.Get().Server.AppID,
			Port:  getPort(config.Get().Server.GRPCAddr),
			Host:  util.ResolveAddr(),
			Metadata: map[string]interface{}{
				"http_port":       getPort(config.Get().Server.HTTPAddr),
				"grpc_port":       getPort(config.Get().Server.GRPCAddr),
				"proxy_http_port": config.Get().Proxy.HTTPPort,
				"proxy_grpc_port": config.Get().Proxy.GRPCPort,
			},
		}); nil != err {
		log.Fatal(err)
	}

	// create message dispatcher.
	if err = loadDispatcher(context.Background()); nil != err {
		log.Fatal(err)
	}

	// initialize search engine.
	if err = search.Init(config.Get().Components.SearchEngine); nil != err {
		log.Fatal(err)
	}

	var coreDao dao.IDao
	if coreDao, err = dao.New(ctx, config.Get().Components.Store, config.Get().Components.Etcd); nil != err {
		log.Fatal(err)
	}

	coreRepo := repository.New(coreDao)
	nodeInstance := runtime.NewNode(context.Background(), newResourceManager(coreRepo), _dispatcher)
	if _apiManager, err = apim.New(context.Background(), coreRepo, _dispatcher); nil != err {
		log.Fatal(err)
	}

	if err = nodeInstance.Start(runtime.NodeConf{Sources: config.Get().Server.Sources}); nil != err {
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
	// initialize ts service.
	_tsSrv.Init(apiManager)
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
	if _tsSrv, err = service.NewTSService(); nil != err {
		log.Fatal(err)
	}
	corev1.RegisterTSHTTPServer(httpSrv.Container, _tsSrv)
}

func serviceRegisterToProxyV1(ctx context.Context, httpSrv *http.Server, grpcSrv *grpc.Server) {
	// register proxy service.
	_proxySrv = service.NewProxyService()
	corev1.RegisterProxyHTTPServer(httpSrv.Container, _proxySrv)
	corev1.RegisterProxyServer(grpcSrv.GetServe(), _proxySrv)
}

func newResourceManager(coreRepo repository.IRepository) types.ResourceManager {
	log.L().Info("create core default resources")
	// default time series.
	tsdbClient := tseries.NewTimeSerier(resource.ParseFrom(config.Get().Components.TimeSeries).Name)
	tsdbClient.Init(resource.ParseFrom(config.Get().Components.TimeSeries))
	return types.NewResources(search.GlobalService, tsdbClient, coreRepo)
}

func loadDispatcher(ctx context.Context) error {
	log.L().Info("load dispatcher...")
	dispatcher := dispatch.New(ctx)
	if err := dispatcher.Start(ctx, config.Get().Dispatcher); nil != err {
		log.L().Error("run dispatcher", zap.Error(err), logger.ID(config.Get().Dispatcher.ID))
		return errors.Wrap(err, "start dispatcher")
	}

	log.L().Info("dispatcher loaded")
	_dispatcher = dispatcher
	return nil
}

func getPort(addr string) int {
	segs := strings.Split(addr, ":")
	p, _ := strconv.Atoi(segs[1])
	return p
}
