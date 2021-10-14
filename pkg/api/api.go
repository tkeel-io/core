package api

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api")

// Registry is api registry.
type Registry struct {
	ctx         context.Context
	services    map[string]IService
	daprService common.Service
}

// NewAPIRegistry returns a new NewAPIRegistry.
func NewAPIRegistry(ctx context.Context, service common.Service) (*Registry, error) {
	return &Registry{
		ctx:         ctx,
		daprService: service,
		services:    make(map[string]IService),
	}, nil
}

// AddService add service to registry.
func (r *Registry) AddService(s IService) error {
	if _, exists := r.services[s.Name()]; exists {
		return fmt.Errorf("service %s aready existed.", s.Name())
	}

	r.services[s.Name()] = s
	return nil
}

func (r *Registry) Start() error {
	// register services.
	for _, s := range r.services {
		if err := s.RegisterService(r.daprService); nil != err {
			return errors.Wrap(err, "register failed")
		}
	}

	return nil
}

func (r *Registry) Close() {}
