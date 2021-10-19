package bootstrap

import (
	"context"
	"fmt"
	"log"

	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/api"
	"github.com/tkeel-io/core/pkg/api/service"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/server"
	daprd "github.com/tkeel-io/core/pkg/service/http"

	"github.com/dapr/go-sdk/service/common"
)

type Server struct {
	conf          *config.Config
	daprService   common.Service
	apiRegistry   *api.Registry
	serverManager *server.Manager
	entityManager *entities.EntityManager
	coroutinePool *ants.Pool

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(ctx context.Context, conf *config.Config) *Server {
	ctx, cancel := context.WithCancel(ctx)

	// create a Dapr service server
	api.SetDefaultPluginID(conf.Server.AppID)
	address := fmt.Sprintf(":%d", conf.Server.AppPort)
	daprService := daprd.NewServiceWithMux(address, api.NewOpenAPIServeMux())

	// create coroutine pool.
	coroutinePool, err := ants.NewPool(conf.Server.CoroutinePoolSize)
	if nil != err {
		log.Fatal(err)
	}

	// create entity manager.
	entityManager := entities.NewEntityManager(ctx, coroutinePool)

	ser := Server{
		ctx:           ctx,
		cancel:        cancel,
		conf:          conf,
		daprService:   daprService,
		coroutinePool: coroutinePool,
		entityManager: entityManager,
		serverManager: server.NewManager(ctx, daprService, conf),
	}

	apiRegistry, err := api.NewAPIRegistry(ctx, daprService)
	if err != nil {
		log.Fatal(err)
	}

	// init api registry.
	if err = initAPIRegistry(ctx, apiRegistry, &conf.APIConfig, entityManager); nil != err {
		log.Fatalf("init ApiRegistry error, %s", err.Error())
	}

	ser.apiRegistry = apiRegistry

	// actor manager init.
	_ = ser.serverManager.Init()

	return &ser
}

func (s *Server) Run() error {
	if err := s.apiRegistry.Start(); nil != err {
		return errors.Unwrap(err)
	}
	if err := s.serverManager.Start(); nil != err {
		return errors.Unwrap(err)
	}

	if err := s.entityManager.Start(); err != nil {
		return errors.Unwrap(err)
	}

	if err := s.daprService.Start(); err != nil {
		return errors.Unwrap(err)
	}

	return nil
}

func (s *Server) Close() {}

func initAPIRegistry(ctx context.Context, apiRegistry *api.Registry, apiConfig *config.APIConfig, entityManager *entities.EntityManager) error {
	var (
		err       error
		eventAPI  *service.EventService
		entityAPI *service.EntityService
	)

	// register event api.
	if eventAPI, err = service.NewEventService(&service.EventServiceConfig{
		RawTopic:          apiConfig.EventAPIConfig.RawTopic,
		TimeSeriesTopic:   apiConfig.EventAPIConfig.TimeSeriesTopic,
		PropertyTopic:     apiConfig.EventAPIConfig.PropertyTopic,
		RelationShipTopic: apiConfig.EventAPIConfig.RelationshipTopic,
		StoreName:         apiConfig.EventAPIConfig.StoreName,
		PubsubName:        apiConfig.EventAPIConfig.PubsubName,
	}); err != nil {
		return errors.Wrap(err, "new event service err")
	}

	if err = apiRegistry.AddService(eventAPI); err != nil {
		return errors.Wrap(err, "api registry add service err")
	}

	if err = apiRegistry.AddService(service.NewTimeSeriesService()); err != nil {
		return errors.Wrap(err, "api registry add service err")
	}

	if entityAPI, err = service.NewEntityService(ctx, entityManager); err != nil {
		return errors.Wrap(err, "new entity service err")
	}

	if err = apiRegistry.AddService(entityAPI); err != nil {
		return errors.Wrap(err, "api registry add service err")
	}

	return nil
}
