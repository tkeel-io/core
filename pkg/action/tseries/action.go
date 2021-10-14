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
