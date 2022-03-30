package dapr

import (
	"context"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
	"github.com/tkeel-io/core/pkg/util/dapr"
	"go.uber.org/zap"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type daprMetadata struct {
	StoreName string `mapstructure:"store_name"`
}

type daprStore struct {
	id        string
	storeName string
}

// Get returns state.
func (d *daprStore) Get(ctx context.Context, key string) (*store.StateItem, error) {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", zfield.Key(key),
			zap.String("store_name", d.storeName), zfield.ID(d.id))
		return nil, errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}

	item, err := conn.GetState(ctx, d.storeName, key)
	if nil != err {
		return nil, errors.Wrap(err, "dapr store get")
	}

	if len(item.Value) == 0 {
		return nil, xerrors.ErrEntityNotFound
	}

	return &store.StateItem{
		Key:      item.Key,
		Etag:     item.Etag,
		Value:    item.Value,
		Metadata: item.Metadata,
	}, nil
}

// Set saves the raw data into store using default state options.
func (d *daprStore) Set(ctx context.Context, key string, data []byte) error {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", zfield.Key(key),
			zap.String("store_name", d.storeName),
			zfield.ID(d.id), zap.String("data", string(data)))
		return errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}
	return errors.Wrap(conn.SaveState(ctx, d.storeName, key, data), "dapr store set")
}

func (d *daprStore) Del(ctx context.Context, key string) error {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", zfield.Key(key),
			zap.String("store_name", d.storeName), zfield.ID(d.id))
		return errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}
	return errors.Wrap(conn.DeleteState(ctx, d.storeName, key), "dapr store del")
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<state.dapr> successful")
	store.Register("dapr", func(properties map[string]interface{}) (store.Store, error) {
		var daprMeta daprMetadata
		if err := mapstructure.Decode(properties, &daprMeta); nil != err {
			return nil, errors.Wrap(err, "decode store.dapr configuration")
		}

		id := util.UUID("sdapr")
		log.L().Info("create store.dapr instance", zfield.ID(id))

		return &daprStore{
			id:        id,
			storeName: daprMeta.StoreName,
		}, nil
	})
}
