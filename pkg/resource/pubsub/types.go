package pubsub

import (
	"context"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, message interface{}) error

type Pubsub interface {
	Send(ctx context.Context, event interface{}) error
	Received(ctx context.Context, receiver MessageHandler) error
	Close() error
}

var registeredPubsubs = make(map[string]PubsubGenerator)

type PubsubGenerator func(map[string]interface{}) (Pubsub, error) // nolint

func Register(name string, handler PubsubGenerator) {
	registeredPubsubs[name] = handler
}

func NewPubsub(metadata resource.Metadata) Pubsub {
	var err error
	var pubsubClient Pubsub
	if generator, has := registeredPubsubs[metadata.Name]; has {
		if pubsubClient, err = generator(metadata.Properties); nil == err {
			log.Debug("new Pubsub instance", zfield.Type(metadata.Name))
			return pubsubClient
		}
		log.Error("new Pubsub instance", zap.Error(err),
			zap.String("name", metadata.Name), zap.Any("properties", metadata.Properties))
	}
	pubsubClient, _ = registeredPubsubs["noop"](metadata.Properties)
	return pubsubClient
}
