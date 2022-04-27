package noop

import (
	"context"
	"github.com/google/uuid"
	"os"
	"sync"

	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/store"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type memStore struct {
	id    string
	store map[string]*store.StateItem
}

var lock = sync.RWMutex{}
var UUID = uuid.New()

func (n *memStore) Get(ctx context.Context, key string) (*store.StateItem, error) {
	lock.RLock()
	defer lock.RUnlock()
	if v, ok := n.store[key]; ok {
		return v, nil
	}
	return nil, xerrors.ErrResourceNotFound
}

// Set saves the raw data into store using default state options.
func (n *memStore) Set(ctx context.Context, key string, data []byte) error {
	lock.Lock()
	defer lock.Unlock()
	n.store[key] = &store.StateItem{
		Key:      key,
		Etag:     UUID.String(),
		Value:    data,
		Metadata: map[string]string{},
	}
	return nil
}

func (n *memStore) Del(ctx context.Context, key string) error {
	lock.Lock()
	defer lock.Unlock()
	delete(n.store, key)
	return nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<state.memory> successful")
	store.Register("memory", func(properties map[string]interface{}) (store.Store, error) {
		id := util.UUID("snoop")
		log.L().Info("create store.noop instance", zfield.ID(id))
		return &memStore{id: id, store: map[string]*store.StateItem{}}, nil
	})
}
