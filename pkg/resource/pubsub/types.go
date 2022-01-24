package pubsub

import (
	"context"

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

type Sender interface {
	Send(ctx context.Context, event interface{}) error
	Close() error
}

type Receiver interface {
	Received(ctx context.Context, receiver MessageHandler) error
	Close() error
}

var registeredPubsubs = make(map[string]Generator)

type Generator func(map[string]interface{}) (Pubsub, error) //

func Register(name string, handler Generator) {
	registeredPubsubs[name] = handler
}

func NewPubsub(metadata resource.Metadata) Pubsub {
	var err error
	var pubsubClient Pubsub
	if generator, has := registeredPubsubs[metadata.Name]; has {
		if pubsubClient, err = generator(metadata.Properties); nil == err {
			return pubsubClient
		}
		log.Error("new Pubsub instance", zap.Error(err),
			zap.String("name", metadata.Name), zap.Any("properties", metadata.Properties))
	}
	pubsubClient, _ = registeredPubsubs["noop"](metadata.Properties)
	return pubsubClient
}
