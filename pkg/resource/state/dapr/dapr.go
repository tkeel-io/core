package dapr

import (
	"context"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/state"
)

type daprMetadata struct {
	StoreName string `mapstructure:"store_name"`
}

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
	state.Register("dapr", func(properties map[string]interface{}) (state.Store, error) {
		var daprMeta daprMetadata
		if err := mapstructure.Decode(properties, &daprMeta); nil != err {
			return nil, errors.Wrap(err, "decode store.dapr configuration")
		}

		daprClient, err := daprSDK.NewClient()
		return &daprStore{
			storeName:  daprMeta.StoreName,
			daprClient: daprClient,
		}, errors.Wrap(err, "new dapr store")
	})
}
