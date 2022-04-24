package dao

import (
	"context"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
)

func (d *Dao) StoreResource(ctx context.Context, res Resource) error {
	var (
		err  error
		key  []byte
		data []byte
	)

	if key, err = res.Codec().Key().Encode(res); nil != err {
		return errors.Wrap(err, "dao store put entity")
	} else if data, err = res.Codec().Value().Encode(res); nil != err {
		return errors.Wrap(err, "dao store entity")
	}

	err = d.stateClient.Set(ctx, string(key), data)
	return errors.Wrap(err, "repo put entity")
}

func (d *Dao) GetStoreResource(ctx context.Context, res Resource) (Resource, error) {
	var (
		err  error
		key  []byte
		item *store.StateItem
	)

	if key, err = res.Codec().Key().Encode(res); nil != err {
		return res, errors.Wrap(err, "dao store get entity")
	}

	if item, err = d.stateClient.Get(ctx, string(key)); nil == err {
		if len(item.Value) == 0 {
			return nil, xerrors.ErrResourceNotFound
		}
		err = res.Codec().Value().Decode(item.Value, res)
	}
	return res, errors.Wrap(err, "dao store get entity")
}

func (d *Dao) RemoveStoreResource(ctx context.Context, res Resource) error {
	key, err := res.Codec().Key().Encode(res)
	if nil != err {
		return errors.Wrap(err, "dao store get entity")
	}

	err = d.stateClient.Del(ctx, string(key))
	return errors.Wrap(err, "dao store del entity")
}
