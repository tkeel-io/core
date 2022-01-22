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
	"time"

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/util"
)

const (
	SMMethodSetConfigs   Method = "SetConfigs"
	SMMethodPatchConfigs Method = "PatchConfigs"
	SMMethodDeleteEntity Method = "DeleteEntity"
)

type Method string

type StateManager interface {
	// start manager.
	Start() error
	// shutdown manager.
	Shutdown() error
	// GetResource return resource manager.
	Resource() ResourceManager
	// route messages cluster.
	RouteMessage(context.Context, MessageContext) error
	// handle message on this node.
	HandleMessage(context.Context, MessageContext) error
}

type ResourceManager interface {
	PubsubClient() pubsub.Pubsub
	SearchClient() *search.Service
	TSeriesClient() tseries.TimeSerier
	Repository() repository.IRepository
}

type StateMachiner interface {
	// GetID return state machine id.
	GetID() string
	// GetStatus returns actor status.
	GetStatus() Status
	// GetEntity returns this.Entity.
	GetEntity() *dao.Entity
	// OnMessage recv message from pubsub.
	OnMessage(ctx Message) bool
	// InvokeMsg dispose entity message.
	HandleLoop()
	// WithContext set actor context.
	WithContext(StateContext) StateMachiner
	// Flush flush entity data.
	Flush(ctx context.Context) error
}

type MessageHandler = func(Message) []WatchKey

type PromiseFunc = func(interface{})

type Message interface {
	Message()
	Promised(interface{})
}

type MessageBase struct {
	startTime      time.Time
	PromiseHandler PromiseFunc `json:"-"`
}

func NewMessageBase(promise PromiseFunc) MessageBase {
	return MessageBase{
		startTime:      time.Now(),
		PromiseHandler: promise,
	}
}

func (ms MessageBase) Message() {}
func (ms MessageBase) Promised(v interface{}) {
	if nil == ms.PromiseHandler {
		return
	}
	ms.PromiseHandler(v)
}

func (ms MessageBase) Elapsed() *util.ElapsedTime {
	return util.NewElapsedFrom(ms.startTime)
}

type WatchKey = mapper.WatchKey

type Status string

type PatchData struct {
	Path     string
	Operator constraint.PatchOperator
	Value    interface{}
}
