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

package runtime

import (
	"context"
	"errors"

	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
)

const EntityStateName = "core-state"
const SubscriptionPrefix = "core.subsc."
const EtcdMapperPrefix = "core.mapper."

type WatchKey = mapper.WatchKey

const (
	StateMarchineTypeBasic        = "BASIC"
	StateMarchineTypeSubscription = "SUBSCRIPTION"
)

var (
	ErrInvalidParams       = errors.New("invalid params")
	ErrInvalidTQLKey       = errors.New("invalid TQL key")
	ErrSubscriptionInvalid = errors.New("invalid subscription")
)

type SMGenerator func(ctx context.Context, base *statem.Base) (statem.StateMarchiner, error)
