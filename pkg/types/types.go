package types

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

type ResourceManager interface {
	Search() *search.Service
	TSDB() tseries.TimeSerier
	Repo() repository.IRepository
	RawData() rawdata.Service
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
}

type IDGenerator interface {
	// returns an uuid.
	ID() string
	// returns an entity id.
	EID() string
	// returns an event id.
	EvID() string
	// returns a requesit id.
	ReqID() string
	// returns a subscription id.
	SubID() string
	// generate id with prefix.
	With(prefix string)
}
