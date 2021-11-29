package statem

import (
	"context"
	"errors"
	"sort"

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	StateFlushPeried = 10

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
	errInvalidJSONPath = errors.New("invalid JSONPath")
)

type StateManager interface {
	Start() error
	SendMsg(msgCtx MessageContext)
	HandleMsg(ctx context.Context, msgCtx MessageContext)
	EscapedEntities(expression string) []string
	SearchFlush(context.Context, map[string]interface{}) error
}

type StateMarchiner interface {
	// GetID return state marchine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	// Setup state marchine setup.
	Setup() error
	// SetConfig set configs.
	SetConfig(map[string]constraint.Config) error
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
	Promised(interface{})
}

type MessageBase struct {
	PromiseHandler PromiseFunc `json:"-"`
}

func (ms MessageBase) Message() {}
func (ms MessageBase) Promised(v interface{}) {
	if nil == ms.PromiseHandler {
		return
	}
	ms.PromiseHandler(v)
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
func (h Header) GetSource() string { return h[MessageCtxHeaderSourceID] }

// SetSource set message source.
func (h Header) SetSource(owner string) { h[MessageCtxHeaderSourceID] = owner }

func (h Header) Get(key string) string { return h[key] }

func (h Header) GetDefault(key, defaultValue string) string {
	if _, has := h[key]; !has {
		return defaultValue
	}
	return h[key]
}

func (h Header) Set(key, value string) { h[key] = value }

type WatchKey = mapper.WatchKey

func SliceAppend(slice sort.StringSlice, vals []string) sort.StringSlice {
	slice = append(slice, vals...)
	return Unique(slice)
}

func Unique(slice sort.StringSlice) sort.StringSlice {
	if slice.Len() <= 1 {
		return slice
	}

	newSlice := sort.StringSlice{slice[0]}

	preVal := slice[0]
	sort.Sort(slice)
	for i := 1; i < slice.Len(); i++ {
		if preVal == slice[i] {
			continue
		}

		preVal = slice[i]
		newSlice = append(newSlice, preVal)
	}
	return newSlice
}
