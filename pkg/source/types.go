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

package source

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

type Type = string

const (
	PubSub Type = "pubsub"
)

// Metadata represents a set of source specific properties.
type Metadata struct {
	Name       string            `json:"name"`
	Type       Type              `json:"type"`
	Properties map[string]string `json:"properties"`
}

type Handler = func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)

type ISource interface {
	String() string
	StartReceiver(fn Handler) error
	Close() error
}

type OpenSourceHandler = func(context.Context, Metadata, common.Service) (ISource, error)

type Generator interface {
	Type() Type
	OpenSource(context.Context, Metadata, common.Service) (ISource, error)
}
