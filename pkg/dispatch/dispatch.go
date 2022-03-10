package dispatch

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/util"
	xkafka "github.com/tkeel-io/core/pkg/util/kafka"
	"github.com/tkeel-io/core/pkg/util/transport"
)

type DispatchConf struct { //nolint
	Name        string
	Upstreams   []string
	Downstreams []string
}

func New(ctx context.Context) *dispatcher { //nolint
	ctx, cancel := context.WithCancel(ctx)
	return &dispatcher{
		id:          util.UUID("dispatcher"),
		ctx:         ctx,
		cancel:      cancel,
		transmitter: transport.New(transport.TransTypeHTTP),
		downstreams: make(map[string]*xkafka.Pubsub),
	}
}

type dispatcher struct {
	id          string
	ctx         context.Context
	cancel      context.CancelFunc
	transmitter transport.Transmitter
	downstreams map[string]*xkafka.Pubsub
}

func (d *dispatcher) Dispatch(ctx context.Context, ev v1.Event) error {
	var err error
	switch ev.Type() {
	case v1.ETCallback:
		err = d.transmitter.Do(ctx, &transport.Request{
			PackageID: ev.ID(),
			Method:    http.MethodPost,
			Address:   ev.CallbackAddr(),
			Header:    ev.Attributes(),
			Payload:   ev.RawData(),
		})
	default:
		return d.dispatch(ctx, ev)
	}

	return errors.Wrap(err, "dispatch event")
}

func (d *dispatcher) Start(ctx context.Context, cfg DispatchConf) error {
	// initialize dispatch upstreams.
	if err := d.initUpstream(ctx, cfg.Upstreams); nil != err {
		return errors.Wrap(err, "init upstream")
	}

	// initialize dispatch downstreams.
	if err := d.initDownstream(ctx, cfg.Downstreams); nil != err {
		return errors.Wrap(err, "init downstream")
	}

	return nil
}

func (d *dispatcher) dispatch(ctx context.Context, ev v1.Event) error {
	eid := ev.Entity()
	info := placement.Global().Select(eid)
	err := d.downstreams[info.ID].Send(ctx, ev)
	return errors.Wrap(err, "dispatch event")
}

func (d *dispatcher) initUpstream(ctx context.Context, streams []string) error {
	return nil
}

func (d *dispatcher) initDownstream(ctx context.Context, streams []string) error {
	for _, stream := range streams {
		streamIns, err := xkafka.NewKafkaPubsub(stream)
		if nil != err {
			return errors.Wrap(err, "create sink instance")
		}
		d.downstreams[streamIns.ID()] = streamIns
		placement.Global().Append(placement.Info{ID: streamIns.ID()})
	}
	return nil
}
