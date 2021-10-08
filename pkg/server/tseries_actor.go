package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tkeel-io/core/pkg/action"
	tseries_action "github.com/tkeel-io/core/pkg/action/tseries"
	batchqueue "github.com/tkeel-io/core/pkg/batch_queue"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/source"
	"github.com/dapr/go-sdk/service/common"
	"go.uber.org/atomic"
)

type TSeriesServer struct {
	name    string
	service common.Service
	sources []source.ISource
	queue   batchqueue.BatchSink
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

func (this *TSeriesServer) Init(serverConfig *config.TSeriesServer) error {

	var (
		err error
	)

	//open sources
	if len(serverConfig.Sources) == 0 {
		return errors.New("server has no source.")
	}

	for _, sourceCfg := range serverConfig.Sources {

		var sourceInst source.ISource
		meta := source.Metadata{Type: sourceCfg.Type, Name: sourceCfg.Name, Properties: sourceCfg.Properties}
		sourceInst, err = source.OpenSource(childContext(this.ctx, this.name), meta, this.service)
		if nil != err {
			return err
		}

		this.sources = append(this.sources, sourceInst)
	}

	//create batch queue.
	queueCfg := serverConfig.BatchQueue
	if this.queue, err = batchqueue.NewBatchSink(childContext(this.ctx, this.name), &batchqueue.BatchQueueConfig{
		Name:                  queueCfg.Name,
		DoSinkFn:              batchHandler,
		MaxBatching:           queueCfg.MaxBatching,
		MaxPendingMessages:    queueCfg.MaxPendingMessages,
		BatchingMaxFlushDelay: time.Millisecond * queueCfg.BatchingMaxFlushDelay,
	}); nil != err {
		return err
	}

	//create action
	this.action = tseries_action.NewAction(childContext(this.ctx, this.name), fmt.Sprintf("%s.action", this.name), this.queue)

	this.ready.Swap(true)

	return nil
}

func (this *TSeriesServer) Run() error {

	if !this.isReady() {
		return errors.New("server not ready.")
	}

	for _, s := range this.sources {
		if err := s.StartReceiver(this.action.Invoke); nil != err {
			return err
		}
	}

	return nil
}

func (this *TSeriesServer) isReady() bool {
	return this.ready.Load()
}

func childContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, "__parent_name__", name)
}

func batchHandler(msgs []interface{}) (err error) {
	fmt.Println("batch handle messages, len(%d)", len(msgs))
	return errors.New("not implement.")
}
