package search

import (
	"context"
	"testing"

	"github.com/tkeel-io/core/pkg/resource/search/driver"

	"github.com/stretchr/testify/assert"
)

func TestGlobalSearch(t *testing.T) {
	assert.Nil(t, GlobalService)
}

func TestService_Use(t *testing.T) {
	service := NewService(nil)
	newServ := service.Use(driver.Elasticsearch)
	assert.Equal(t, newServ, service)
}

func TestService_With(t *testing.T) {
	service := NewService(nil)
	newServ := service.With(driver.Elasticsearch)
	assert.NotEqual(t, newServ, service)
}

func TestNewService(t *testing.T) {
	fake := driver.Type("fake")
	engine := &fakeEngine{}
	registered := map[driver.Type]driver.SearchEngine{fake: engine}
	NewService(registered)
}

func TestService_Register(t *testing.T) {
	service := NewService(nil)
	assert.Nil(t, service.drivers)
	var fake driver.Type = "fake"
	engine := fakeEngine{}
	service.Register(fake, engine)
	assert.NotNil(t, service.drivers)
	d, ok := service.drivers[fake]
	assert.True(t, ok)
	assert.Equal(t, engine, d)
}

type fakeEngine struct{}

func (f fakeEngine) BuildIndex(ctx context.Context, index, content string) error {
	return nil
}

func (f fakeEngine) Search(ctx context.Context, request driver.SearchRequest) (driver.SearchResponse, error) {
	return driver.SearchResponse{}, nil
}

func (f fakeEngine) Delete(ctx context.Context, id string) error {
	return nil
}
