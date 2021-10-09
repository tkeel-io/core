package api

import (
	"context"
	"fmt"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api")

// APIRegistry is api registry
type APIRegistry struct {
	ctx         context.Context
	services    map[string]IService
	daprService common.Service
}

// NewAPIRegistry returns a new NewAPIRegistry
func NewAPIRegistry(ctx context.Context, service common.Service) (*APIRegistry, error) {

	return &APIRegistry{
		ctx:         ctx,
		daprService: service,
		services:    make(map[string]IService),
	}, nil
}

// AddService add service to registry
func (this *APIRegistry) AddService(s IService) error {

	if _, exists := this.services[s.Name()]; exists {
		return fmt.Errorf("service %s aready existed.", s.Name())
	}

	this.services[s.Name()] = s
	return nil
}

// Start start
func (this *APIRegistry) Start() error {

	//register services.
	for _, s := range this.services {
		if err := s.RegisterService(this.daprService); nil != err {
			return err
		}
	}

	return nil
}

// Close
func (this *APIRegistry) Close() {}
