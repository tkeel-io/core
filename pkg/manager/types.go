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

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository"
)

const CoreAPISender = "core.api"

type TemplateEntityID struct{}

var (
	ErrMapperTQLInvalid    = errors.New("invalid TQL")
	ErrEntityNotFound      = errors.New("not found")
	ErrEntityAreadyExisted = errors.New("entity already existed")
)

type APIManager interface {
	// OnRespond handle message.
	OnRespond(context.Context, *holder.Response)
	// CreateEntity create entity.
	CreateEntity(context.Context, *Base) (*BaseRet, error)
	// UpdateEntity update entity.
	PatchEntity(context.Context, *Base, []*v1.PatchData, ...Option) (*BaseRet, []byte, error)
	// DeleteEntity delete entity.
	DeleteEntity(context.Context, *Base) error
	// GetProperties returns entity properties.
	GetEntity(context.Context, *Base) (*BaseRet, error)
	// AppendMapper append entity mapper.
	AppendMapper(context.Context, *mapper.Mapper) error
	AppendMapperZ(context.Context, *mapper.Mapper) error

	// Expression.
	AppendExpression(context.Context, []repository.Expression) error
	RemoveExpression(context.Context, []repository.Expression) error
	GetExpression(context.Context, repository.Expression) (*repository.Expression, error)
	ListExpression(context.Context, *Base) ([]*repository.Expression, error)
}

type Metadata map[string]string

type Option func(meta Metadata)

func NewPathConstructorOption(pc v1.PathConstructor) Option {
	return func(meta Metadata) {
		meta[v1.MetaPathConstructor] = string(pc)
	}
}
