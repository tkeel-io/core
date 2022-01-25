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
	"errors"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/kit/log"
)

type WatchKey = mapper.WatchKey

const (
	SMTypeBasic        = "BASIC"
	SMTypeSubscription = "SUBSCRIPTION"

	// state machine required fileds.
	SMFieldType     = "type"
	SMFieldOwner    = "owner"
	SMFieldSource   = "source"
	SMFieldTemplate = "template"
)

var (
	ErrInvalidParams       = errors.New("invalid params")
	ErrInvalidTQLKey       = errors.New("invalid TQL key")
	ErrSubscriptionInvalid = errors.New("invalid subscription")
)

func unwrapString(str string) string {
	if len(str) > 2 {
		return str[1 : len(str)-1]
	}
	log.Warn("unwrap string failed", zfield.Value(str))
	return str
}
