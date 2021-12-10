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
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
)

var (
	log = logger.NewLogger("core.entities")

	errEntityNotFound      = errors.New("entity not found")
	errEmptyEntityMapper   = errors.New("empty entity mapper")
	errSubscriptionInvalid = errors.New("invalid params")
)

const (
	MessageCtxHeaderEntityType = "x-entity-type"

	EntityTypeBaseEntity   = "base"
	EntityTypeSubscription = "subscription"
)

type EntityOp interface {
	statem.StateMarchiner
}

type EntitySubscriptionOp interface {
	EntityOp

	GetMode() string
}

type WatchKey = mapper.WatchKey
