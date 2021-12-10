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

package tseries

import (
	"context"
	"errors"

	batchq "github.com/tkeel-io/core/pkg/batch_queue"
	"github.com/tkeel-io/core/pkg/logger"

	"github.com/dapr/go-sdk/service/common"
)

var log = logger.NewLogger("kcore.action")

// Action is a time-series action.
type Action struct {
	name  string
	queue batchq.BatchSink
	ctx   context.Context
}

// NewAction returns a new time-series action.
func NewAction(ctx context.Context, name string, queue batchq.BatchSink) *Action {
	return &Action{
		ctx:   ctx,
		name:  name,
		queue: queue,
	}
}

// Invoke handle input messages.
func (action *Action) Invoke(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	if nil == e {
		return false, errors.New("empty request")
	}

	if nil == e.Data {
		log.Warnf("recv empty message, PubsubName:%s, Topic:%s, ID:%s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	}

	// decode data, then send data to queue.
	_ = action.queue.Send(action.ctx, e.Data)
	return false, nil
}
