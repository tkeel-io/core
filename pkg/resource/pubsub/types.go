package pubsub

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type EventHandler func(context.Context, v1.Event) error

type Pubsub interface {
	Commiter

	ID() string
	Send(context.Context, v1.Event) error
	Received(context.Context, EventHandler) error
	Close() error
}

type Sender interface {
	ID() string
	Send(context.Context, v1.Event) error
	Close() error
}

type Receiver interface {
	Commiter

	ID() string
	Received(context.Context, EventHandler) error
	Close() error
}

type Commiter interface {
	ID() string
	Commit(v interface{}) error
}

var registeredPubsubs = make(map[string]Generator)

type Generator func(string, map[string]interface{}) (Pubsub, error) //

func Register(name string, handler Generator) {
	registeredPubsubs[name] = handler
}

func NewPubsub(id string, metadata resource.Metadata) Pubsub {
	var err error
	var pubsubClient Pubsub
	if generator, has := registeredPubsubs[metadata.Name]; has {
		if pubsubClient, err = generator(id, metadata.Properties); nil == err {
			return pubsubClient
		}
		log.L().Error("new Pubsub instance", zap.Error(err),
			zap.String("name", metadata.Name), zap.Any("properties", metadata.Properties))
	}
	pubsubClient, _ = registeredPubsubs["noop"](id, metadata.Properties)
	return pubsubClient
}
