package driver

import (
	"context"

	"github.com/tkeel-io/kit/log"
)

const DriverNameNoop Type = "noop"

type noopSearchEngine struct{}

func NewNoopSearchEngine(_ map[string]interface{}) (SearchEngine, error) {
	return &noopSearchEngine{}, nil
}

func (ns *noopSearchEngine) BuildIndex(ctx context.Context, index, content string) error {
	log.L().Error("BuildIndex noop")
	return nil
}
func (ns *noopSearchEngine) Search(ctx context.Context, request SearchRequest) (SearchResponse, error) {
	log.L().Error("Search noop")
	return SearchResponse{}, nil
}
func (ns *noopSearchEngine) Delete(ctx context.Context, id string) error {
	log.L().Error("Delete noop")
	return nil
}

func NoopDriver() Type {
	return DriverNameNoop
}

func init() {
	registerDrivers[DriverNameNoop] = NewNoopSearchEngine
}
