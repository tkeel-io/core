package noop

import (
	"context"
	"os"

	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource/store"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type noopStore struct {
	id string
}

func (n *noopStore) Get(ctx context.Context, key string) (*store.StateItem, error) {
	return nil, xerrors.ErrResourceNotFound
}

// Set saves the raw data into store using default state options.
func (n *noopStore) Set(ctx context.Context, key string, data []byte) error {
	return nil
}

func (n *noopStore) Del(ctx context.Context, key string) error {
	return nil
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Resource<state.noop> successful")
	store.Register("noop", func(properties map[string]interface{}) (store.Store, error) {
		id := util.UUID("snoop")
		log.L().Info("create store.noop instance", logf.ID(id))
		return &noopStore{id: id}, nil
	})
}
