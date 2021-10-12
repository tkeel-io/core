package entities

import (
	"errors"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
)

var (
	log = logger.NewLogger("core.entities")

	errMapperExisted = errors.New("mapper aready exists.")
	errEntityExisted = errors.New("entity aready exists.")
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
	InvokeMsg(entityId string, values map[string]interface{})
	// SetMapper
	SetMapper(m mapper.Mapper) error
	// GetMapper returns a mapper.
	GetMapper(mid string) mapper.Mapper
	// GetMappers
	GetMappers() []mapper.Mapper
	// TentacleModify notify tentacle event.
	TentacleModify(requestId, entityId string)
	// GetTentacles returns tentacles.
	GetTentacles(requestId, entityId string)
}
