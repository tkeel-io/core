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

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/statem"
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
	OnMessage(context.Context, statem.MessageContext) error
	// CreateEntity create entity.
	CreateEntity(ctx context.Context, base *Base) (*Base, error)
	// DeleteEntity delete entity.
	DeleteEntity(ctx context.Context, en *Base) (base *Base, err error)
	// GetProperties returns entity properties.
	GetProperties(ctx context.Context, en *Base) (base *Base, err error)
	// SetProperties set entity properties.
	SetProperties(ctx context.Context, en *Base) (base *Base, err error)
	// PatchEntity patch entity properties.
	PatchEntity(ctx context.Context, en *Base, patchData []*pb.PatchData) (base *Base, err error)
	// AppendMapper append entity mapper.
	AppendMapper(ctx context.Context, en *Base) (base *Base, err error)
	// RemoveMapper remove entity mapper.
	RemoveMapper(ctx context.Context, en *Base) (base *Base, err error)
	// CheckSubscription check subscription.
	CheckSubscription(ctx context.Context, en *Base) (err error)
	// SetConfigs set entity configs.
	SetConfigs(ctx context.Context, en *Base) (base *Base, err error)
	// PatchConfigs patch entity configs.
	PatchConfigs(ctx context.Context, en *Base, patchData []*statem.PatchData) (base *Base, err error)
	// QueryConfigs returns entity configs.
	QueryConfigs(ctx context.Context, en *Base, propertyIDs []string) (base *Base, err error)
}

func convert(en *dao.Entity) *Base {
	return &Base{
		ID:         en.ID,
		Type:       en.Type,
		Owner:      en.Owner,
		Source:     en.Source,
		Version:    en.Version,
		LastTime:   en.LastTime,
		TemplateID: en.TemplateID,
		Properties: en.Properties,
		ConfigFile: en.ConfigFile,
	}
}
