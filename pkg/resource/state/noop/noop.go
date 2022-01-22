package noop

import (
	"context"
	"errors"

	"github.com/tkeel-io/core/pkg/resource/state"
)

type noopStore struct{}

func (n *noopStore) Get(ctx context.Context, key string) (*state.StateItem, error) {
	return nil, errors.New("noop store")
}

// Set saves the raw data into store using default state options.
func (n *noopStore) Set(ctx context.Context, key string, data []byte) error {
	return nil
}

func (n *noopStore) Del(ctx context.Context, key string) error {
	return nil
}

func init() {
	state.Register("noop", func(properties map[string]interface{}) (state.Store, error) {
		return &noopStore{}, nil
	})
}
