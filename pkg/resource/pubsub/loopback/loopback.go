package loopback

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

type EventFunc func(context.Context, v1.Event) error

type Loopback struct {
	id           string
	eventHandler EventFunc
}

func NewLoopback() *Loopback {
	return &Loopback{id: "loopback"}
}

func (d *Loopback) ID() string {
	return d.id
}

func (d *Loopback) Send(ctx context.Context, event v1.Event) error {
	log.Debug("pubsub.loopback send", zfield.Message(event), zfield.ID(d.id))
	err := d.eventHandler(ctx, event)
	return errors.Wrap(err, "send event")
}

func (d *Loopback) Received(ctx context.Context, handler EventFunc) error {
	log.Info("pubsub.loopback start receive message", zfield.ID(d.id))
	d.eventHandler = handler
	return nil
}

func (d *Loopback) Commit(v interface{}) error {
	return nil
}

func (d *Loopback) Close() error {
	log.Info("pubsub.loopback close", zfield.ID(d.id))
	return nil
}
