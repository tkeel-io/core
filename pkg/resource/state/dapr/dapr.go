package dapr

import (
	"context"
	"os"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/state"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type daprMetadata struct {
	StoreName string `mapstructure:"store_name"`
}

type daprStore struct {
	id         string
	storeName  string
	daprClient daprSDK.Client
}

// Get returns state.
func (d *daprStore) Get(ctx context.Context, key string) (*state.StateItem, error) {
	item, err := d.daprClient.GetState(ctx, d.storeName, key)
	if nil != err {
		if len(item.Value) == 0 {
			return nil, xerrors.ErrEntityNotFound
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
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<state.dapr> successful")
	state.Register("dapr", func(properties map[string]interface{}) (state.Store, error) {
		var daprMeta daprMetadata
		if err := mapstructure.Decode(properties, &daprMeta); nil != err {
			return nil, errors.Wrap(err, "decode store.dapr configuration")
		}

		id := util.UUID()
		log.Info("create store.dapr instance", zfield.ID(id))

		daprClient, err := daprSDK.NewClient()
		return &daprStore{
			id:         id,
			storeName:  daprMeta.StoreName,
			daprClient: daprClient,
		}, errors.Wrap(err, "new dapr store")
	})
}
