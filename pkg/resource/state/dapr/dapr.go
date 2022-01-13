package dapr

import (
	"context"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/state"
)

type daprStore struct {
	daprClient daprSDK.Client
}

func (d *daprStore) Get(ctx context.Context, storeName, key string) (*state.StateItem, error) {
	item, err := d.daprClient.GetState(ctx, storeName, key)
	if nil != err {
		return nil, errors.Wrap(err, "dapr store get")
	}
	return &state.StateItem{
		Key:      item.Key,
		Etag:     item.Etag,
		Value:    item.Value,
		Metadata: item.Metadata,
	}, nil
}

// SaveState saves the raw data into store using default state options.
func (d *daprStore) Set(ctx context.Context, storeName, key string, data []byte) error {
	return errors.Wrap(d.daprClient.SaveState(ctx, storeName, key, data), "dapr store set")
}

func init() {
	state.Register("noop", func() (state.Store, error) {
		daprClient, err := daprSDK.NewClient()
		return &daprStore{daprClient: daprClient}, errors.Wrap(err, "new dapr store")
	})
}
