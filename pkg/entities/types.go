package entities

import (
	"errors"

	"github.com/tkeel-io/core/pkg/logger"
)

const (
	EntityCtxHeaderOwner     = "x-owner"
	EntityCtxHeaderType      = "x-type"
	EntityCtxHeaderSourceID  = "x-source"
	EntityCtxHeaderTargetID  = "x-target"
	EntityCtxHeaderRequestID = "x-reqsuest_id"
	EntityCtxHeaderPluginID  = "x-plugin"

	TentacleOperatorAppend = "append"
	TentacleOperatorRemove = "remove"
)

var (
	log = logger.NewLogger("core.entities")

	errEntityNotFound = errors.New("entity not found")
)

type EntityOp interface {
	// GetID returns entity id.
	GetID() string
	// GetProperty returns entity property.
	GetProperty(string) interface{}
	// GetAllProperties returns entity properties.
	GetAllProperties() *EntityBase
	// SetProperties set entity properties
	SetProperties(*EntityBase) (*EntityBase, error)
	// DeleteProperty delete entity property.
	DeleteProperty(string) error
	// OnMessage recv message from pubsub.
	OnMessage(ctx EntityContext) bool
	// InvokeMsg dispose entity message.
	InvokeMsg()
	// SetMapper set mapper into entity.
	SetMapper(m MapperDesc) error
	// GetMapper returns a mapper.
	GetMapper(mid string) (MapperDesc, error)
	// GetMappers
	GetMappers() []MapperDesc
}

type EntitySubscriptionOp interface {
	EntityOp

	GetMode() string
}

type EntityContext struct {
	Headers Header
	Message Message
}

func NewEntityContext(msg Message) EntityContext {
	return EntityContext{
		Headers: Header{},
		Message: msg,
	}
}

type Header map[string]string

func (h Header) GetTargetID() string {
	return h[EntityCtxHeaderTargetID]
}

func (h Header) SetTargetID(targetID string) {
	h[EntityCtxHeaderTargetID] = targetID
}

func (h Header) GetOwner() string {
	return h[EntityCtxHeaderOwner]
}

func (h Header) SetOwner(owner string) {
	h[EntityCtxHeaderOwner] = owner
}

func (h Header) GetPluginID() string {
	return h[EntityCtxHeaderPluginID]
}

func (h Header) SetPluginID(plugin string) {
	h[EntityCtxHeaderPluginID] = plugin
}

func (h Header) GetEntityType() string {
	t, has := h[EntityCtxHeaderType]
	if !has {
		t = EntityTypeDevice
	}
	return t
}

func (h Header) SetEntityType(plugin string) {
	h[EntityCtxHeaderType] = plugin
}

type PromiseFunc = func(interface{})

type Message interface {
	Message()
	Promise() PromiseFunc
}

type messageBase struct{}

func (ms *messageBase) Message() {}
func (ms *messageBase) Promise() PromiseFunc {
	return func(interface{}) {
		// do nothing...
	}
}

type AttacheHandler = func()
