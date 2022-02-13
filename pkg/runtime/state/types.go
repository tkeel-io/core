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

package state

import (
	"context"

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/message"
)

type Machiner interface {
	// GetID return state machine id.
	GetID() string
	// GetStatus returns actor status.
	GetStatus() Status
	// GetEntity returns this.Entity.
	GetEntity() *dao.Entity
	// OnMessage recv message from pubsub.
	Invoke(ctx message.Context) error
	// Flush flush entity data.
	Flush(ctx context.Context) error
}

type WatchKey = mapper.WatchKey

type Status string

type PatchData struct {
	Path     string                   `json:"path"`
	Operator constraint.PatchOperator `json:"operator"`
	Value    interface{}              `json:"value"`
}

type MessageHandler = func(message.Context) []WatchKey
