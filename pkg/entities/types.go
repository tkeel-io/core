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

package entities

import (
	"context"
	"errors"

	cloudevents "github.com/cloudevents/sdk-go"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/runtime/state"
)

const CoreAPISender = "core.api"

type TemplateEntityID struct{}

var (
	ErrMapperTQLInvalid    = errors.New("invalid TQL")
	ErrEntityNotFound      = errors.New("not found")
	ErrEntityAreadyExisted = errors.New("entity already existed")
)

type EntityManager interface {
	// Start start Entity manager.
	Start() error
	// OnMessage handle message.
	OnMessage(context.Context, cloudevents.Event) error
	// CreateEntity create entity.
	CreateEntity(context.Context, *Base) (*Base, error)
	// DeleteEntity delete entity.
	DeleteEntity(context.Context, *Base) error
	// GetProperties returns entity properties.
	GetProperties(context.Context, *Base) (*Base, error)
	// SetProperties set entity properties.
	SetProperties(context.Context, *Base) (*Base, error)
	// PatchEntity patch entity properties.
	PatchEntity(context.Context, *Base, []*pb.PatchData) (*Base, error)
	// AppendMapper append entity mapper.
	AppendMapper(context.Context, *Base) (*Base, error)
	// RemoveMapper remove entity mapper.
	RemoveMapper(context.Context, *Base) (*Base, error)
	// CheckSubscription check subscription.
	CheckSubscription(context.Context, *Base) error
	// SetConfigs set entity configs.
	SetConfigs(context.Context, *Base) (*Base, error)
	// PatchConfigs patch entity configs.
	PatchConfigs(context.Context, *Base, []*state.PatchData) (*Base, error)
	// QueryConfigs returns entity configs.
	QueryConfigs(context.Context, *Base, []string) (*Base, error)
}
