package types

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

type ResourceManager interface {
	Pubsub() pubsub.Pubsub
	Search() *search.Service
	TSDB() tseries.TimeSerier
	Repo() repository.IRepository
}

type Republisher interface {
	RouteMessage(ctx context.Context, ev cloudevents.Event) error
}

// state machine manager.
type Manager interface {
	// start manager.
	Start() error
	// shutdown manager.
	Shutdown() error
	// GetResource return resource manager.
	Resource() ResourceManager
	// route messages cluster.
	RouteMessage(context.Context, cloudevents.Event) error
	// handle message on this node.
	SetRepublisher(republisher Republisher)
}
