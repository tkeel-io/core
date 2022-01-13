package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/service/mock"
)

func Test_NewSearchService(t *testing.T) {
	sv := NewSearchService(mock.NewSearchMock())
	assert.NotNil(t, sv)
}
