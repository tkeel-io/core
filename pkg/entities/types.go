package entities

import (
	"errors"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	EntityCtxHeaderSourceID  = "x-source"
	EntityCtxHeaderTargetID  = "x-target"
	EntityCtxHeaderRequestID = "x-request-id"
)

var (
	log = logger.NewLogger("core.entities")

	errEntityExisted  = errors.New("entity already exists")
	errEntityNotFound = errors.New("entity not found")
)

type EntityOp interface {
	// GetID returns entity id.
	GetID() string
	// GetProperty returns entity property.
	GetProperty(string) interface{}
	// SetProperty set entity property.
	SetProperty(string, interface{}) error
	// GetAllProperties returns entity properties.
	GetAllProperties() map[string]interface{}
	// SetProperties set entity properties
	SetProperties(map[string]interface{}) error
	// DeleteProperty delete entity property.
	DeleteProperty(string) error
	InvokeMsg(EntityContext)
	SetMapper(m mapper.Mapper) error
	// GetMapper returns a mapper.
	GetMapper(mid string) mapper.Mapper
	// GetMappers
	GetMappers() []mapper.Mapper
}

type EntityContext struct {
	Headers Header
	Message Message
}

func (ec *EntityContext) TargetID() string {
	return ec.Headers[EntityCtxHeaderTargetID]
}

func (ec *EntityContext) SetTarget(targetID string) {
	ec.Headers[EntityCtxHeaderTargetID] = targetID
}

type Header map[string]string

type Message interface {
	Message()
}
