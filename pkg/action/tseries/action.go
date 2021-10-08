package tseries

import (
	"context"
	"errors"

	batchqueue "github.com/tkeel-io/core/pkg/batch_queue"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/dapr/go-sdk/service/common"
)

var log = logger.NewLogger("kcore.action")

// TSeriesAction is a time-series action.
type TSeriesAction struct {
	name  string
	queue batchqueue.BatchSink
	ctx   context.Context
}

//NewAction returns a new time-series action.
func NewAction(ctx context.Context, name string, queue batchqueue.BatchSink) *TSeriesAction {
	return &TSeriesAction{
		ctx:   ctx,
		name:  name,
		queue: queue,
	}
}

// Invoke handle input messages.
func (tsa *TSeriesAction) Invoke(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {

	if nil == e {
		return false, errors.New("empty request.")
	}

	if nil == e.Data {
		log.Warnf("recv empty message, PubsubName:%s, Topic:%s, ID:%s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	}

	//decode data, then send data to queue.
	tsa.queue.Send(tsa.ctx, e.Data)
	return false, nil
}
