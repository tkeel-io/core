package dapr

import (
	"context"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/state"
)

type daprStore struct {
	storeName  string
	daprClient daprSDK.Client
}

// Get returns state.
func (d *daprStore) Get(ctx context.Context, key string) (*state.StateItem, error) {
	item, err := d.daprClient.GetState(ctx, d.storeName, key)
	if nil != err {
		if false {
			// TODO: 将dapr-state的error转换成Core的error.
			err = xerrors.ErrEntityNotFound
		}
		return nil, errors.Wrap(err, "dapr store get")
	}
	return &state.StateItem{
		Key:      item.Key,
		Etag:     item.Etag,
		Value:    item.Value,
		Metadata: item.Metadata,
	}, nil
}

// Set saves the raw data into store using default state options.
func (d *daprStore) Set(ctx context.Context, key string, data []byte) error {
	return errors.Wrap(d.daprClient.SaveState(ctx, d.storeName, key, data), "dapr store set")
}

func (d *daprStore) Del(ctx context.Context, key string) error {
	return errors.Wrap(d.daprClient.DeleteState(ctx, d.storeName, key), "dapr store del")
}

func init() {
	state.Register("dapr", func(storeName string) (state.Store, error) {
		daprClient, err := daprSDK.NewClient()
		return &daprStore{daprClient: daprClient}, errors.Wrap(err, "new dapr store")
	})
}
