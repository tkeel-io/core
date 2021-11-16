package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/core/pkg/server"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"

	// User import
	Entity_v1 "github.com/tkeel-io/core/api/core/v1"

	openapi "github.com/tkeel-io/core/api/openapi/v1"
)

var (
	// app name
	Name string
	// http addr
	HTTPAddr string
	// grpc addr
	GRPCAddr string
)

func init() {
	flag.StringVar(&Name, "name", "core", "app name.")
	flag.StringVar(&HTTPAddr, "http_addr", ":31234", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
}

func main() {
	flag.Parse()

	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	app := app.New(Name,
		&log.Conf{
			App:   Name,
			Level: "debug",
			Dev:   true,
		},
		serverList...,
	)

	{ // User service
		EntitySrv := service.NewEntityService()
		Entity_v1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)
		Entity_v1.RegisterEntityServer(grpcSrv.GetServe(), EntitySrv)

		OpenapiSrv := service.NewOpenapiService()
		openapi.RegisterOpenapiHTTPServer(httpSrv.Container, OpenapiSrv)
		openapi.RegisterOpenapiServer(grpcSrv.GetServe(), OpenapiSrv)
	}

	if err := app.Run(context.TODO()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err := app.Stop(context.TODO()); err != nil {
		panic(err)
	}
}
