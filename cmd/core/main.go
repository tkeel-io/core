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
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/search"
	"github.com/tkeel-io/core/pkg/server"
	"github.com/tkeel-io/core/pkg/service"

	"github.com/panjf2000/ants/v2"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"
)

var (
	Name          string
	HTTPAddr      string
	GRPCAddr      string
	SearchBrokers string
)

func init() {
	flag.StringVar(&Name, "name", "core", "app name.")
	flag.StringVar(&HTTPAddr, "http_addr", ":6789", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
	flag.StringVar(&SearchBrokers, "search_brokers", "http://localhost:9200", "search brokers address.")
}

func main() {
	flag.Parse()

	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	coreApp := app.New(Name,
		&log.Conf{
			App:   Name,
			Level: "debug",
			Dev:   true,
		},
		serverList...,
	)

	coroutinePool, err := ants.NewPool(100)
	if nil != err {
		log.Fatal(err)
	}

	searchClient := search.NewESClient(strings.Split(SearchBrokers, ",")...)
	entityManager, err := entities.NewEntityManager(context.Background(), coroutinePool, searchClient)
	if nil != err {
		log.Fatal(err)
	}

	{
		// User service
		// create coroutine pool.

		EntitySrv, err := service.NewEntityService(context.Background(), entityManager, searchClient)
		if nil != err {
			log.Fatal(err)
		}
		Core_v1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)
		Core_v1.RegisterEntityServer(grpcSrv.GetServe(), EntitySrv)

		SubscriptionSrv, err := service.NewSubscriptionService(context.Background(), entityManager)
		if nil != err {
			log.Fatal(err)
		}
		Core_v1.RegisterSubscriptionHTTPServer(httpSrv.Container, SubscriptionSrv)
		Core_v1.RegisterSubscriptionServer(grpcSrv.GetServe(), SubscriptionSrv)

		TopicSrv, err := service.NewTopicService(context.Background(), entityManager)
		if nil != err {
			log.Fatal(err)
		}
		Core_v1.RegisterTopicHTTPServer(httpSrv.Container, TopicSrv)
		Core_v1.RegisterTopicServer(grpcSrv.GetServe(), TopicSrv)

		SearchSrv := service.NewSearchService(searchClient)
		Core_v1.RegisterSearchHTTPServer(httpSrv.Container, SearchSrv)
		Core_v1.RegisterSearchServer(grpcSrv.GetServe(), SearchSrv)
	}

	// run.
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
