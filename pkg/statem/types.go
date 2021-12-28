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
	"sort"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/resource/tseries"
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
)

type StateManager interface {
	Start() error
	SendMsg(msgCtx MessageContext)
	GetDaprClient() dapr.Client
	HandleMsg(ctx context.Context, msgCtx MessageContext)
	EscapedEntities(expression string) []string
	SearchFlush(context.Context, map[string]interface{}) error
	TimeSeriesFlush(context.Context, []tseries.TSeriesData) error
}

type StateMarchiner interface {
	// GetID return state marchine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	// Setup state marchine setup.
	Setup() error
	// SetStatus set state-marchine status.
	SetStatus(Status)
	// GetStatus returns state-marchine status.
	GetStatus() Status
	// SetConfig set configs.
	SetConfig(map[string]constraint.Config) error
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

type Status string

const (
	SMStatusActive   Status = "active"
	SMStatusInactive Status = "inactive"
	SMStatusDeleted  Status = "deleted"
)
