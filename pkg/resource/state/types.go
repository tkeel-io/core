package state

import (
	"context"

	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// StateItem represents a single state item.
type StateItem struct { //nolint
	Key      string
	Value    []byte
	Etag     string
	Metadata map[string]string
}

type Store interface {
	// GetState retrieves state from specific store using default consistency option.
	Get(ctx context.Context, storeName, key string) (item *StateItem, err error)
	// SaveState saves the raw data into store using default state options.
	Set(ctx context.Context, storeName, key string, data []byte) error
}

var registeredStores = make(map[string]StoreGenerator)

type StoreGenerator func() (Store, error)

func Register(name string, handler StoreGenerator) {
	registeredStores[name] = handler
}

func NewStore(name string) Store {
	var err error
	if generator, has := registeredStores[name]; has {
		var s Store
		if s, err = generator(); nil == err {
			return s
		}
		log.Error("Generate store", zap.String("name", name), zap.Error(err))
	}
	s, _ := registeredStores["noop"]()
	return s
}
