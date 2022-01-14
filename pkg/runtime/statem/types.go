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

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	MessageCtxHeaderOwner     = "x-owner"
	MessageCtxHeaderType      = "x-type"
	MessageCtxHeaderSourceID  = "x-source"
	MessageCtxHeaderTargetID  = "x-target"
	MessageCtxHeaderStateType = "x-state-type"
	MessageCtxHeaderRequestID = "x-reqsuest-id"
	MessageCtxHeaderChannelID = "x-channel-id"
)

type StateManager interface {
	Start() error
	Shutdown() error
	RouteMessage(context.Context, MessageContext) error
	HandleMessage(context.Context, MessageContext) error
}

type StateMachiner interface {
	// GetID return state machine id.
	GetID() string
	// GetBase returns state.Base
	GetBase() *Base
	GetStatus() Status
	// OnMessage recv message from pubsub.
	OnMessage(ctx Message) bool
	// InvokeMsg dispose entity message.
	HandleLoop()
	WithContext(StateContext) StateMachiner
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
