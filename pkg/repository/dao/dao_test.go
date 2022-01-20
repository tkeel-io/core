package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/constraint"
)

func TestDao(t *testing.T) {
	d := New(context.Background(), "core-entity", &storeMock{})
	assert.NotNil(t, d, "result not nil")
}

func TestDao_Put(t *testing.T) {
	d := New(context.Background(), "core-entity", &storeMock{})
	assert.NotNil(t, d, "result not nil")
	err := d.Put(context.TODO(), &Entity{
		ID:         "device123",
		Type:       "DEVICE",
		Owner:      "admin",
		Source:     "dm",
		Version:    0,
		Properties: map[string]constraint.Node{"temp": constraint.NewNode(25)},
	})
	assert.Nil(t, err, "nil error")
}

func TestDao_Get(t *testing.T) {
	d := New(context.Background(), "core-entity", &storeMock{})
	assert.NotNil(t, d, "result not nil")
	en, err := d.Get(context.TODO(), "device123")
	assert.Nil(t, err, "nil error")
	assert.Equal(t, "device123", en.ID)
}

func TestDao_Del(t *testing.T) {
	d := New(context.Background(), "core-entity", &storeMock{})
	assert.NotNil(t, d, "result not nil")
	err := d.Del(context.TODO(), "device123")
	assert.Nil(t, err, "nil error")
}
