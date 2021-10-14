package entities

import (
	"errors"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	EntityCtxHeaderSourceId  = "x-source"
	EntityCtxHeaderTargetId  = "x-target"
	EntityCtxHeaderRequestId = "x-request-id"
)

var (
	log = logger.NewLogger("core.entities")

	errMapperExisted  = errors.New("mapper aready exists.")
	errEntityExisted  = errors.New("entity aready exists.")
	errEntityNotFound = errors.New("entity not found.")
)

type EntityOp interface {
	//GetId returns entity id.
	GetId() string
	// GetProperty returns entity property.
	GetProperty(string) interface{}
	// SetProperty set entity property.
	SetProperty(string, interface{}) error
	//GetAllProperties returns entity properties.
	GetAllProperties() map[string]interface{}
	// SetProperties set entity properties
	SetProperties(map[string]interface{}) error
	// DeleteProperty delete entity property.
	DeleteProperty(string) error
	// InvokeMsg
	InvokeMsg(EntityContext)
	// SetMapper
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

func (ec *EntityContext) TargetId() string {
	return ec.Headers[EntityCtxHeaderTargetId]
}

func (ec *EntityContext) SetTarget(targetId string) {
	ec.Headers[EntityCtxHeaderTargetId] = targetId
}

type Header map[string]string

type Message interface {
	Message()
}
