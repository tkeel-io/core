package pubsub

import (
	"context"
	"net/url"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
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

type Generator func(string, string) (Pubsub, error)

func Register(name string, handler Generator) {
	registeredPubsubs[name] = handler
}

func NewPubsub(id string, urlText string) Pubsub {
	var err error
	var pubsubClient Pubsub
	pubsubClient, _ = registeredPubsubs["noop"](id, urlText)

	if id != "" {
		id = util.UUID("pubsub")
	}

	// parse url.
	urlIns, err := url.Parse(urlText)
	if nil != err {
		log.L().Error("parse url", logf.Error(err), logf.URL(urlText))
		return pubsubClient
	}

	// new pubsub instance.
	if generator, has := registeredPubsubs[urlIns.Scheme]; has {
		var pubsubClient0 Pubsub
		if pubsubClient0, err = generator(id, urlText); nil == err {
			return pubsubClient0
		}
		log.L().Error("new Pubsub instance", logf.Error(err), logf.URL(urlText))
	}

	return pubsubClient
}
