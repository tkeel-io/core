package types

import (
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

type resourceManager struct {
	defaultSearch  *search.Service
	defaultTSDB    tseries.TimeSerier
	defaultRepo    repository.IRepository
	defaultRawData rawdata.Service
}

func NewResources(
	searchClient *search.Service,
	tseriesClient tseries.TimeSerier,
	rawdataClient rawdata.Service,
	repoClient repository.IRepository) ResourceManager {
	return &resourceManager{
		defaultRepo:    repoClient,
		defaultTSDB:    tseriesClient,
		defaultSearch:  searchClient,
		defaultRawData: rawdataClient,
	}
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

func (r *resourceManager) RawData() rawdata.Service {
	return r.defaultRawData
}
