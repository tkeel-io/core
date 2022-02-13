package loopback

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

type loopbackPubsub struct {
	id           string
	eventHandler pubsub.EventHandler
}

func (d *loopbackPubsub) Send(ctx context.Context, event cloudevents.Event) error {
	log.Debug("pubsub.loopback send", zfield.Message(event), zfield.ID(d.id))
	err := d.eventHandler(ctx, event)
	return errors.Wrap(err, "send event")
}

func (d *loopbackPubsub) Received(ctx context.Context, receiver pubsub.EventHandler) error {
	log.Info("pubsub.loopback start receive message", zfield.ID(d.id))
	d.eventHandler = receiver
	return nil
}

func (d *loopbackPubsub) Commit(v interface{}) error {
	return nil
}

func (d *loopbackPubsub) Close() error {
	log.Info("pubsub.loopback close", zfield.ID(d.id))
	return nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.loopback> successful")
	pubsub.Register("loopback", func(map[string]interface{}) (pubsub.Pubsub, error) {
		id := util.UUID()
		log.Info("create pubsub.loopback instance", zfield.ID(id))
		return &loopbackPubsub{id: id}, nil
	})
}
