/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statem

import (
	"context"
	"errors"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/runtime/environment"
)

const (
	StateFlushPeried = 10

	MessageCtxHeaderOwner     = "x-owner"
	MessageCtxHeaderType      = "x-type"
	MessageCtxHeaderSourceID  = "x-source"
	MessageCtxHeaderTargetID  = "x-target"
	MessageCtxHeaderStateType = "x-state-type"
	MessageCtxHeaderRequestID = "x-reqsuest-id"
	MessageCtxHeaderChannelID = "x-channel-id"

	MapperOperatorAppend   = "append"
	MapperOperatorRemove   = "remove"
	TentacleOperatorAppend = "append"
	TentacleOperatorRemove = "remove"
)

var (
	errInvalidMapperOp   = errors.New("invalid mapper operator")
	errInvalidJSONPath   = errors.New("invalid JSONPath")
	ErrInvalidProperties = errors.New("statem invalid properties")
	ErrPropertyNotFound  = errors.New("property not found")
)

type StateManagerV2 interface {
	Start() error
	Shutdown()
	HandleMessage(MessageContext) error
}

type StateManager interface {
	Start() error
	SendMsg(msgCtx MessageContext)
	GetDaprClient() dapr.Client
	HandleMsg(ctx context.Context, msgCtx MessageContext)
	EscapedEntities(expression string) []string
	SearchFlush(context.Context, map[string]interface{}) error
	TimeSeriesFlush(context.Context, []tseries.TSeriesData) error
	SetConfigs(context.Context, *Base) error
	PatchConfigs(context.Context, *Base, []*PatchData) error
	AppendConfigs(context.Context, *Base) error
	RemoveConfigs(context.Context, *Base, []string) error
}

type StateMachinerV2 interface {
	// GetID return state machine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	// OnMessage recv message from pubsub.
	OnMessage(ctx Message) bool
	// Flush flush entity data.
	Flush(ctx context.Context) error
}

type StateMachiner interface {
	// GetID return state machine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	// Setup state machine setup.
	Setup() error
	// SetStatus set state-machine status.
	SetStatus(Status)
	// GetStatus returns state-machine status.
	GetStatus() Status
	// SetConfig set entity configs.
	SetConfigs(map[string]constraint.Config) error
	// PatchConfigs patch configs.
	PatchConfigs(patchDatas []*PatchData) error
	// AppendConfig append entity property config.
	AppendConfigs(map[string]constraint.Config) error
	// RemoveConfig remove entity property configs.
	RemoveConfigs(propertyIDs []string) error
	// LoadEnvironments load environments.
	LoadEnvironments(environment.ActorEnv)
	// OnMessage recv message from pubsub.
	OnMessage(ctx Message) bool
	// InvokeMsg dispose entity message.
	HandleLoop()
	// StateManager returns state manager.
	GetManager() StateManager
	// Flush flush entity data.
	Flush(ctx context.Context) error
}

type Flusher interface {
	FlushState() error
	FlushSearch() error
	FlushTimeSeries() error
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

type WatchKey = mapper.WatchKey

type Status string

type PatchData struct {
	Path     string
	Operator constraint.PatchOperator
	Value    interface{}
}

//----------------- mock.
type Pubsub interface {
}
