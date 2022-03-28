package noop

import (
	"context"
	"os"

	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/kit/log"
)

type noopPubsub struct {
	id string
}

func (d *noopPubsub) ID() string {
	return d.id
}

func (d *noopPubsub) Send(ctx context.Context, event v1.Event) error {
	log.L().Debug("pubsub.noop send", zfield.Message(event), zfield.ID(d.id))
	return nil
}

func (d *noopPubsub) Received(ctx context.Context, receiver pubsub.EventHandler) error {
	log.L().Info("pubsub.noop start receive message", zfield.ID(d.id))
	return nil
}

func (d *noopPubsub) Commit(v interface{}) error {
	return nil
}

func (d *noopPubsub) Close() error {
	log.L().Info("pubsub.noop close", zfield.ID(d.id))
	return nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.noop> successful")
	pubsub.Register("noop", func(id string, urlText string) (pubsub.Pubsub, error) {
		log.L().Info("create pubsub.noop instance", zfield.ID(id))
		return &noopPubsub{id: id}, nil
	})
}
