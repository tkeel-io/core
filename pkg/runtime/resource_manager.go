package runtime

import (
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/types"
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
	repoClient repository.IRepository) types.ResourceManager {
	return &resourceManager{
		defaultRepo:   repoClient,
		defaultTSDB:   tseriesClient,
		defaultPubsub: pubsubClient,
		defaultSearch: searchClient,
	}
}

func (r *resourceManager) Pubsub() pubsub.Pubsub {
	return r.defaultPubsub
}

func (r *resourceManager) Search() *search.Service {
	return r.defaultSearch
}

func (r *resourceManager) TSDB() tseries.TimeSerier {
	return r.defaultTSDB
}

func (r *resourceManager) Repo() repository.IRepository {
	return r.defaultRepo
}
