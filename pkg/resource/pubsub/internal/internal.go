package internal

import (
	"context"
	"os"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type internalPubsub struct {
	id           string
	eventHandler pubsub.EventHandler
}

func (d *internalPubsub) Send(ctx context.Context, event cloudevents.Event) error {
	log.Debug("pubsub.internal send", zfield.Message(event), zfield.ID(d.id))
	err := d.eventHandler(ctx, event)
	return errors.Wrap(err, "send event")
}

func (d *internalPubsub) Received(ctx context.Context, receiver pubsub.EventHandler) error {
	log.Info("pubsub.internal start receive message", zfield.ID(d.id))
	d.eventHandler = receiver
	return nil
}

func (d *internalPubsub) Commit(v interface{}) error {
	return nil
}

func (d *internalPubsub) Close() error {
	log.Info("pubsub.internal close", zfield.ID(d.id))
	return nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.internal> successful")
	pubsub.Register("internal", func(map[string]interface{}) (pubsub.Pubsub, error) {
		id := util.UUID()
		log.Info("create pubsub.internal instance", zfield.ID(id))
		return &internalPubsub{id: id}, nil
	})
}
