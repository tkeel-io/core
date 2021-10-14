package bootstrap

import (
	"context"
	"fmt"
	"log"

	ants "github.com/panjf2000/ants/v2"
	"github.com/tkeel-io/core/pkg/api"
	"github.com/tkeel-io/core/pkg/api/service"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/server"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/tkeel-io/core/pkg/service/http"
	//daprd "github.com/dapr/go-sdk/service/http"
)

type Server struct {
	conf          *config.Config
	daprService   common.Service
	apiRegistry   *api.APIRegistry
	serverManager *server.ServerManager
	entityManager *entities.EntityManager
	coroutinePool *ants.Pool

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(ctx context.Context, conf *config.Config) *Server {

	ctx, cancel := context.WithCancel(ctx)

	// create a Dapr service server
	api.SetDefaultPluginId(conf.Server.AppId)
	address := fmt.Sprintf(":%d", conf.Server.AppPort)
	daprService := daprd.NewServiceWithMux(address, api.NewOpenApiServeMux())

	ser := Server{
		ctx:           ctx,
		cancel:        cancel,
		conf:          conf,
		daprService:   daprService,
		serverManager: server.NewServerManager(ctx, daprService, conf),
	}

	//create a api registry.
	apiRegistry, err := api.NewAPIRegistry(ctx, daprService)
	if nil != err {
		log.Fatal(err)
	}

	//init api registry.
	if err = initApiRegistry(apiRegistry, &conf.ApiConfig); nil != err {
		log.Fatalf("init ApiRegistry error, %s", err.Error())
	}

	ser.apiRegistry = apiRegistry

	//actor manager init.
	_ = ser.serverManager.Init()

	return &ser
}

func (this *Server) Run() error {
	var err error
	if err = this.apiRegistry.Start(); nil != err {
		return err
	} else if err = this.serverManager.Start(); nil != err {
		return err
	}
	return this.daprService.Start()
}

func (this *Server) Close() {}

func initApiRegistry(apiRegistry *api.APIRegistry, apiConfig *config.ApiConfig) error {

	var (
		err       error
		eventApi  *service.EventService
		entityApi *service.EntityService
	)

	// register event api.
	if eventApi, err = service.NewEventService(&service.EventServiceConfig{
		RawTopic:          apiConfig.EventApiConfig.RawTopic,
		TimeSeriesTopic:   apiConfig.EventApiConfig.TimeSeriesTopic,
		PropertyTopic:     apiConfig.EventApiConfig.PropertyTopic,
		RelationShipTopic: apiConfig.EventApiConfig.RelationShipTopic,
		StoreName:         apiConfig.EventApiConfig.StoreName,
		PubsubName:        apiConfig.EventApiConfig.PubsubName,
	}); nil != err {
		return err
	}

	if err = apiRegistry.AddService(eventApi); nil != err {
		return err
	}

	//register time-series api.
	if err = apiRegistry.AddService(service.NewTimeSeriesService()); nil != err {
		return err
	}

	//init entity api
	if entityApi, err = service.NewEntityService(&service.EntityServiceConfig{
		TableName:   apiConfig.EntityApiConfig.TableName,
		StateName:   apiConfig.EntityApiConfig.StateName,
		BindingName: apiConfig.EntityApiConfig.BindingName,
	}); nil != err {
		return err
	}

	if err = apiRegistry.AddService(entityApi); nil != err {
		return err
	}

	return err
}
