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

package manager

import (
	"context"
	"errors"

	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/runtime/state"
)

const CoreAPISender = "core.api"

type TemplateEntityID struct{}

var (
	ErrMapperTQLInvalid    = errors.New("invalid TQL")
	ErrEntityNotFound      = errors.New("not found")
	ErrEntityAreadyExisted = errors.New("entity already existed")
)

type APIManager interface {
	// Start start Entity manager.
	Start() error
	// OnRespond handle message.
	OnRespond(context.Context, *holder.Response)
	// CreateEntity create entity.
	UpdateEntity(context.Context, *Base) (*Base, error)
	// UpdateEntity update entity.
	CreateEntity(context.Context, *Base) (*Base, error)
	// DeleteEntity delete entity.
	DeleteEntity(context.Context, *Base) error
	// GetProperties returns entity properties.
	GetEntity(context.Context, *Base) (*Base, error)
	// SetProperties set entity properties.
	UpdateEntityProps(context.Context, *Base) (*Base, error)
	// PatchEntity patch entity properties.
	PatchEntityProps(context.Context, *Base, []state.PatchData) (*Base, error)
	// GetEntityProps returns entity configs.
	GetEntityProps(context.Context, *Base, []string) (*Base, error)
	// SetConfigs set entity configs.
	UpdateEntityConfigs(context.Context, *Base) (*Base, error)
	// PatchConfigs patch entity configs.
	PatchEntityConfigs(context.Context, *Base, []state.PatchData) (*Base, error)
	// GetEntityConfigs returns entity configs.
	GetEntityConfigs(context.Context, *Base, []string) (*Base, error)
	// AppendMapper append entity mapper.
	AppendMapper(context.Context, *Base) error
	// RemoveMapper remove entity mapper.
	RemoveMapper(context.Context, *Base) error
	// CheckSubscription check subscription.
	CheckSubscription(context.Context, *Base) error
}
