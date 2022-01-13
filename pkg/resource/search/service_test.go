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
	newServ := service.Use(driver.WithElasticsearch)
	assert.NotEqual(t, newServ, service)
	assert.Equal(t, 1, len(newServ.selectOpts))
}

func TestService_SetDefaultSelectOptions(t *testing.T) {
	service := NewService(nil)
	newServ := service.SetSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, newServ, service)
	assert.Equal(t, 1, len(newServ.selectOpts))
}

func TestService_AppendSelectOptions(t *testing.T) {
	service := NewService(nil)
	newServ := service.SetSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, newServ, service)
	assert.Equal(t, 1, len(newServ.selectOpts))

	newServ.AppendSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, 2, len(newServ.selectOpts))
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
