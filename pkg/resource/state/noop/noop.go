package noop

import (
	"context"
	"errors"
	"os"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/state"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type noopStore struct {
	id string
}

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
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<state.noop> successful")
	state.Register("noop", func(properties map[string]interface{}) (state.Store, error) {
		id := util.UUID()
		log.Info("create store.noop instance", zfield.ID(id))
		return &noopStore{id: id}, nil
	})
}
