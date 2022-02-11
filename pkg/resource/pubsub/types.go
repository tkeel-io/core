package pubsub

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type EventHandler func(context.Context, cloudevents.Event) error

type Pubsub interface {
	Commiter
	Send(context.Context, cloudevents.Event) error
	Received(context.Context, EventHandler) error
	Close() error
}

type Sender interface {
	Send(context.Context, cloudevents.Event) error
	Close() error
}

type Receiver interface {
	Commiter
	Received(context.Context, EventHandler) error
	Close() error
}

type Commiter interface {
	Commit(v interface{}) error
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
