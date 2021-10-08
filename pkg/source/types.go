package source

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

type SourceType = string

const (
	SourceTypePubSub SourceType = "pubsub"
)

// Metadata represents a set of source specific properties
type Metadata struct {
	Name       string            `json:"name"`
	Type       SourceType        `json:"type"`
	Properties map[string]string `json:"properties"`
}

type SourceHandler = func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)

type ISource interface {
	String() string
	StartReceiver(fn SourceHandler) error
	Close() error
}

type OpenSourceHandler = func(context.Context, Metadata, common.Service) (ISource, error)

type SourceGenerator interface {
	Type() SourceType
	OpenSource(context.Context, Metadata, common.Service) (ISource, error)
}
