package server

import (
	"context"
	"fmt"
	"os"

	batchq "github.com/tkeel-io/core/pkg/batch_queue"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/print"
	"github.com/tkeel-io/core/pkg/source"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/action"
	tseriesaction "github.com/tkeel-io/core/pkg/action/tseries"
	"go.uber.org/atomic"
)

type TSeriesServer struct {
	name    string
	service common.Service
	sources []source.ISource
	queue   batchq.BatchSink
	action  action.IAction

	ready atomic.Bool
	ctx   context.Context
}

func NewTSeriesServer(ctx context.Context, name string, service common.Service) *TSeriesServer {
	return &TSeriesServer{
		ctx:     ctx,
		name:    name,
		service: service,
	}
}

func (t *TSeriesServer) Init(serverConfig *config.TSeriesServer) error {
	var (
		err error
	)

	if len(serverConfig.Sources) == 0 {
		return errors.New("server has no source")
	}

	for _, sourceCfg := range serverConfig.Sources {
		// open sources.
		var sourceInst source.ISource
		meta := source.Metadata{Type: sourceCfg.Type, Name: sourceCfg.Name, Properties: sourceCfg.Properties}
		sourceInst, err = source.OpenSource(childContext(t.ctx, t.name), meta, t.service)
		if nil != err {
			return errors.Wrap(err, "open source err")
		}

		t.sources = append(t.sources, sourceInst)
	}

	// create batch queue.
	queueCfg := serverConfig.BatchQueue
	t.queue, err = batchq.NewBatchSink(childContext(t.ctx, t.name), &batchq.Config{
		Name:                  queueCfg.Name,
		DoSinkFn:              batchHandler,
		MaxBatching:           queueCfg.MaxBatching,
		MaxPendingMessages:    queueCfg.MaxPendingMessages,
		BatchingMaxFlushDelay: queueCfg.BatchingMaxFlushDelay,
	})
	if err != nil {
		return errors.Unwrap(err)
	}

	// create action.
	t.action = tseriesaction.NewAction(childContext(t.ctx, t.name), fmt.Sprintf("%s.action", t.name), t.queue)

	t.ready.Swap(true)

	return nil
}

func (t *TSeriesServer) Run() error {
	if !t.isReady() {
		return errors.New("server not ready")
	}

	for _, s := range t.sources {
		if err := s.StartReceiver(t.action.Invoke); nil != err {
			return errors.Unwrap(err)
		}
	}

	return nil
}

func (t *TSeriesServer) isReady() bool {
	return t.ready.Load()
}

func childContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ContextKey("__parent_name__"), name)
}

func batchHandler(msgs []interface{}) (err error) {
	print.FailureStatusEvent(os.Stderr, "batch handle messages, len(%d)", len(msgs))
	return errors.New("not implement")
}

type ContextKey string
