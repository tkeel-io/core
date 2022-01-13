package search

import (
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
	newServ := service.SetDefaultSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, newServ, service)
	assert.Equal(t, 1, len(newServ.selectOpts))
}

func TestService_AppendSelectOptions(t *testing.T) {
	service := NewService(nil)
	newServ := service.SetDefaultSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, newServ, service)
	assert.Equal(t, 1, len(newServ.selectOpts))

	newServ.AppendSelectOptions(driver.WithElasticsearch)
	assert.Equal(t, 2, len(newServ.selectOpts))
}
