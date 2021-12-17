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
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	Core_v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/print"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/search"
	"github.com/tkeel-io/core/pkg/server"
	"github.com/tkeel-io/core/pkg/service"

	"github.com/panjf2000/ants/v2"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/influxdb"
	_ "github.com/tkeel-io/core/pkg/resource/tseries/noop"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"
)

var (
	cfgFile       string
	HTTPAddr      string
	GRPCAddr      string
	SearchBrokers string
)

func init() {
	flag.StringVar(&cfgFile, "conf", "config.yml", "config file path.")
	flag.StringVar(&HTTPAddr, "http_addr", ":6789", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
	flag.StringVar(&SearchBrokers, "search_brokers", "http://localhost:9200", "search brokers address.")
}

func main() {
	flag.Parse()
	// load configuration.
	config.InitConfig(cfgFile)

	// new servers.
	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	coreApp := app.New(config.GetConfig().Server.AppID,
		&log.Conf{
			App:    config.GetConfig().Server.AppID,
			Level:  config.GetConfig().Logger.Level,
			Dev:    config.GetConfig().Logger.Dev,
			Output: config.GetConfig().Logger.Output,
		},
		serverList...,
	)

	var entityManager *entities.EntityManager
	searchClient := search.NewESClient(strings.Split(SearchBrokers, ",")...)

	// create coroutine pool.
	if coroutinePool, err := ants.NewPool(5000); nil != err {
		log.Fatal(err)
	} else if stateManager, err := runtime.NewManager(context.Background(), coroutinePool, searchClient); nil != err {
		log.Fatal(err)
	} else if entityManager, err = entities.NewEntityManager(context.Background(), stateManager, searchClient); nil != err {
		log.Fatal(err)
	}

	{
		// User service
		if EntitySrv, err := service.NewEntityService(context.Background(), entityManager, searchClient); nil != err {
			log.Fatal(err)
		} else if SubscriptionSrv, err := service.NewSubscriptionService(context.Background(), entityManager); nil != err {
			log.Fatal(err)
		} else if TopicSrv, err := service.NewTopicService(context.Background(), entityManager); nil != err {
			log.Fatal(err)
		} else {
			// register entity service.
			Core_v1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)
			Core_v1.RegisterEntityServer(grpcSrv.GetServe(), EntitySrv)

			// register topic service.
			Core_v1.RegisterTopicHTTPServer(httpSrv.Container, TopicSrv)
			Core_v1.RegisterTopicServer(grpcSrv.GetServe(), TopicSrv)

			// register subscription service.
			Core_v1.RegisterSubscriptionHTTPServer(httpSrv.Container, SubscriptionSrv)
			Core_v1.RegisterSubscriptionServer(grpcSrv.GetServe(), SubscriptionSrv)

			// register search service.

			SearchSrv := service.NewSearchService(searchClient)
			Core_v1.RegisterSearchHTTPServer(httpSrv.Container, SearchSrv)
			Core_v1.RegisterSearchServer(grpcSrv.GetServe(), SearchSrv)
			print.SuccessStatusEvent(os.Stdout, "all seavice registered.")
		}
	}

	// running...
	print.SuccessStatusEvent(os.Stdout, "everything is ready for execution.")
	if err := entityManager.Start(); nil != err {
		panic(err)
	} else if err = coreApp.Run(context.TODO()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err := coreApp.Stop(context.TODO()); err != nil {
		panic(err)
	}
}
