package runtime

import (
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/runtime/state"
)

type resourceManager struct {
	defaultPubsub pubsub.Pubsub
	defaultSearch *search.Service
	defaultTSDB   tseries.TimeSerier
	defaultRepo   repository.IRepository
}

func NewResources(
	pubsubClient pubsub.Pubsub,
	searchClient *search.Service,
	tseriesClient tseries.TimeSerier,
	repoClient repository.IRepository) state.ResourceManager {
	return &resourceManager{
		defaultRepo:   repoClient,
		defaultTSDB:   tseriesClient,
		defaultPubsub: pubsubClient,
		defaultSearch: searchClient,
	}
}

func (r *resourceManager) PubsubClient() pubsub.Pubsub {
	return r.defaultPubsub
}

func (r *resourceManager) SearchClient() *search.Service {
	return r.defaultSearch
}

func (r *resourceManager) TSeriesClient() tseries.TimeSerier {
	return r.defaultTSDB
}

func (r *resourceManager) Repository() repository.IRepository {
	return r.defaultRepo
}
