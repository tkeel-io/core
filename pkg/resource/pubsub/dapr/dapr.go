package dapr

import (
	"context"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/kit/log"
)

type daprMetadata struct {
	TopicName    string `json:"topic_name" mapstructure:"topic_name"`
	PubsubName   string `json:"pubsub_name" mapstructure:"pubsub_name"`
	ConsumerType string `json:"consumer_type" mapstructure:"consumer_type"`
}

type daprPubsub struct {
	id           string
	topicName    string
	pubsubName   string
	consumerType string
}

func (d *daprPubsub) ID() string {
	return d.id
}

func (d *daprPubsub) Send(ctx context.Context, event v1.Event) error {
	panic("never used")
}

func (d *daprPubsub) Received(ctx context.Context, handler pubsub.EventHandler) error {
	log.Debug("pubsub.dapr start receive message", zfield.ID(d.id))
	Register(&Consumer{id: d.id, handler: handler})
	return errors.Wrap(nil, "register message handler")
}

func (d *daprPubsub) Commit(v interface{}) error {
	return nil
}

func (d *daprPubsub) Close() error {
	log.Debug("pubsub.dapr close", zfield.ID(d.id))
	Unregister(&Consumer{id: d.id})
	return errors.Wrap(nil, "unregister message handler")
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.dapr> successful")
	pubsub.Register("dapr", func(id string, properties map[string]interface{}) (pubsub.Pubsub, error) {
		var daprMeta daprMetadata
		if err := mapstructure.Decode(properties, &daprMeta); nil != err {
			return nil, errors.Wrap(err, "decode pubsub.dapr configuration")
		}

		log.Info("create pubsub.dapr instance", zfield.ID(id))

		return &daprPubsub{
			id:           id,
			topicName:    daprMeta.TopicName,
			pubsubName:   daprMeta.PubsubName,
			consumerType: daprMeta.ConsumerType,
		}, nil
	})
}
