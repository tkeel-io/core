package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {
	ns := &noopStore{}
	_, err := ns.Get(context.Background(), "device123")
	assert.NotNil(t, err, "noop GET errors")
}

func Test_Set(t *testing.T) {
	ns := &noopStore{}
	err := ns.Set(context.Background(), "device123", []byte(""))
	assert.Nil(t, err, "noop set")
}
