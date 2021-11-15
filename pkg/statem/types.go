package statem

import (
	"context"
	"errors"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	MessageCtxHeaderOwner     = "x-owner"
	MessageCtxHeaderSourceID  = "x-source"
	MessageCtxHeaderTargetID  = "x-target"
	MessageCtxHeaderStateType = "x-state-type"
	MessageCtxHeaderRequestID = "x-reqsuest_id"

	MapperOperatorAppend   = "append"
	MapperOperatorRemove   = "remove"
	TentacleOperatorAppend = "append"
	TentacleOperatorRemove = "remove"
)

var (
	log = logger.NewLogger("core.state-marchine")

	errInvalidMapperOp = errors.New("invalid mapper operator")
)

type StateManager interface {
	Start() error
	SendMsg(msgCtx MessageContext)
	HandleMsg(ctx context.Context, msgCtx MessageContext)
	EscapedEntities(expression string) []string
}

type StateMarchiner interface {
	// GetID return state marchine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	// OnMessage recv message from pubsub.
	OnMessage(ctx Message) bool
	// InvokeMsg dispose entity message.
	HandleLoop()
	// StateManager returns state manager.
	GetManager() StateManager
}

type MessageHandler = func(Message) []WatchKey

type PromiseFunc = func(interface{})

type Message interface {
	Message()
	Promise() PromiseFunc
}

type messageBase struct {
	PromiseHandler PromiseFunc `json:"-"`
}

func (ms messageBase) Message() {}
func (ms messageBase) Promise() PromiseFunc {
	if nil == ms.PromiseHandler {
		return func(interface{}) {}
	}
	return ms.PromiseHandler
}

type Header map[string]string

type MessageContext struct {
	Headers Header
	Message Message
}

// GetTargetID returns message target id.
func (h Header) GetTargetID() string { return h[MessageCtxHeaderTargetID] }

// SetTargetID set target state marchine id.
func (h Header) SetTargetID(targetID string) { h[MessageCtxHeaderTargetID] = targetID }

// GetOwner returns message owner.
func (h Header) GetOwner() string { return h[MessageCtxHeaderOwner] }

// SetOwner set message owner.
func (h Header) SetOwner(owner string) { h[MessageCtxHeaderOwner] = owner }

// GetSource returns message source field.
func (h Header) GetSource() string { return h[MessageCtxHeaderOwner] }

// SetSource set message source.
func (h Header) SetSource(owner string) { h[MessageCtxHeaderOwner] = owner }

func (h Header) Get(key string) string { return h[key] }

func (h Header) GetDefault(key, defaultValue string) string {
	if _, has := h[key]; !has {
		return defaultValue
	}
	return h[key]
}

func (h Header) Set(key, value string) { h[key] = value }

type WatchKey = mapper.WatchKey
