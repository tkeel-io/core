package source

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

type Type = string

const (
	PubSub Type = "pubsub"
)

// Metadata represents a set of source specific properties.
type Metadata struct {
	Name       string            `json:"name"`
	Type       Type              `json:"type"`
	Properties map[string]string `json:"properties"`
}

type Handler = func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)

type ISource interface {
	String() string
	StartReceiver(fn Handler) error
	Close() error
}

type OpenSourceHandler = func(context.Context, Metadata, common.Service) (ISource, error)

type Generator interface {
	Type() Type
	OpenSource(context.Context, Metadata, common.Service) (ISource, error)
}
