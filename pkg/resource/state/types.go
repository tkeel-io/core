package state

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource"
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
	Get(ctx context.Context, key string) (item *StateItem, err error)
	// SaveState saves the raw data into store using default state options.
	Set(ctx context.Context, key string, data []byte) error
	// Del delete record from store.
	Del(ctx context.Context, key string) error
}

var registeredStores = make(map[string]StoreGenerator)

type StoreGenerator func(map[string]interface{}) (Store, error)

func Register(name string, handler StoreGenerator) {
	registeredStores[name] = handler
}

func NewStore(metadata resource.Metadata) Store {
	var err error
	var storeClient Store
	if generator, has := registeredStores[metadata.Name]; has {
		if storeClient, err = generator(metadata.Properties); nil == err {
			return storeClient
		}
		log.Error("new Store instance", zap.Error(err),
			zap.String("name", metadata.Name), zap.Any("properties", metadata.Properties))
	}
	storeClient, _ = registeredStores["noop"](metadata.Properties)
	return storeClient
}
