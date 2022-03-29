package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewSearchService(t *testing.T) {
	sv := NewSearchService()
	assert.NotNil(t, sv)
}
