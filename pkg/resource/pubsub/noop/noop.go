package noop

import (
	"context"
	"os"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/logfield"
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
	log.L().Debug("[+]pubsub.noop send", logf.Message(event), logf.ID(d.id))
	return nil
}

func (d *noopPubsub) Received(ctx context.Context, receiver pubsub.EventHandler) error {
	log.L().Info("[+]pubsub.noop start receive message", logf.ID(d.id))
	return nil
}

func (d *noopPubsub) Commit(v interface{}) error {
	return nil
}

func (d *noopPubsub) Close() error {
	log.L().Info("pubsub.noop close", logf.ID(d.id))
	return nil
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.noop> successful")
	pubsub.Register("noop", func(id string, urlText string) (pubsub.Pubsub, error) {
		log.L().Info("create pubsub.noop instance", logf.ID(id))
		return &noopPubsub{id: id}, nil
	})
}
