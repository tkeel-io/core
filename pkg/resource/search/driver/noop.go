package driver

import "context"

const DriverNameNoop Type = "noop"

type noopSearchEngine struct{}

func NewNoopSearchEngine(_ map[string]interface{}) (SearchEngine, error) {
	return &noopSearchEngine{}, nil
}

func (ns *noopSearchEngine) BuildIndex(ctx context.Context, index, content string) error {
	return nil
}
func (ns *noopSearchEngine) Search(ctx context.Context, request SearchRequest) (SearchResponse, error) {
	return SearchResponse{}, nil
}
func (ns *noopSearchEngine) Delete(ctx context.Context, id string) error {
	return nil
}

func NoopDriver() Type {
	return DriverNameNoop
}

func init() {
	registerDrivers[DriverNameNoop] = NewNoopSearchEngine
}
