package noop

import (
	"context"
	"os"

	cloudevents "github.com/cloudevents/sdk-go"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type noopPubsub struct {
	id string
}

func (d *noopPubsub) Send(ctx context.Context, event cloudevents.Event) error {
	log.Debug("pubsub.noop send", zfield.Message(event), zfield.ID(d.id))
	return nil
}

func (d *noopPubsub) Received(ctx context.Context, receiver pubsub.EventHandler) error {
	log.Info("pubsub.noop start receive message", zfield.ID(d.id))
	return nil
}

func (d *noopPubsub) Commit(v interface{}) error {
	return nil
}

func (d *noopPubsub) Close() error {
	log.Info("pubsub.noop close", zfield.ID(d.id))
	return nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.noop> successful")
	pubsub.Register("noop", func(map[string]interface{}) (pubsub.Pubsub, error) {
		id := util.UUID()
		log.Info("create pubsub.noop instance", zfield.ID(id))
		return &noopPubsub{id: id}, nil
	})
}
