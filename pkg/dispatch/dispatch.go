package dispatch

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/resource/pubsub/loopback"
	"github.com/tkeel-io/core/pkg/util"
	xkafka "github.com/tkeel-io/core/pkg/util/kafka"
	"github.com/tkeel-io/core/pkg/util/transport"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

/*

	TODO:
		1. dispatch 模块输入：
			a. loopback
			b. dapr pubsubDispatchConf

*/

type DispatchConf struct {
	Sinks []string
}

func New(ctx context.Context) *dispatcher { //nolint
	ctx, cancel := context.WithCancel(ctx)
	return &dispatcher{
		id:          util.UUID("dispatcher"),
		ctx:         ctx,
		cancel:      cancel,
		transmitter: transport.New(transport.TransTypeHTTP),
		sinks:       make(map[string]*xkafka.KafkaPubsub),
	}
}

type dispatcher struct {
	id          string
	ctx         context.Context
	cancel      context.CancelFunc
	transmitter transport.Transmitter
	sinks       map[string]*xkafka.KafkaPubsub
	loopback    *loopback.Loopback
}

func (d *dispatcher) Dispatch(ctx context.Context, ev v1.Event) error {
	var err error
	switch ev.Type() {
	case v1.ETCallback:
		err = d.transmitter.Do(ctx, &transport.Request{
			PackageID: ev.ID(),
			Method:    http.MethodPost,
			Address:   ev.CallbackAddr(),
			Payload:   ev.RawData(),
		})
	default:
		if err = d.loopback.Send(ctx, ev); nil != err {
			log.Error("dispatch event", zap.Error(err),
				zap.Any("event", ev), zfield.DispatcherID(d.id))
		}
	}

	return errors.Wrap(err, "dispatch event")

}

func (d *dispatcher) Start(cfg DispatchConf) error {
	// start loopback.
	d.loopback = loopback.NewLoopback()
	d.loopback.Received(d.ctx, func(ctx context.Context, e v1.Event) error {
		eid := e.Entity()
		info := placement.Global().Select(eid)
		return d.sinks[info.ID].Send(ctx, e)
	})
	// TODO: start dapr source.

	if err := d.initSinks(context.Background(), cfg.Sinks); nil != err {
		return errors.Wrap(err, "init sinks")
	}

	return nil
}

func (d *dispatcher) initSinks(ctx context.Context, sinks []string) error {
	for _, sink := range sinks {
		sinkIns, err := xkafka.NewKafkaPubsub(sink)
		if nil != err {
			return errors.Wrap(err, "create sink instance")
		}
		d.sinks[sinkIns.ID()] = sinkIns
		placement.Global().Append(placement.Info{ID: sinkIns.ID()})
	}
	return nil
}
