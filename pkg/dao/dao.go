package dao

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/state"
)

type Dao struct {
	storeName   string
	stateClient state.Store
}

func New(ctx context.Context, storeName string, stateClient state.Store) *Dao {
	return &Dao{stateClient: stateClient, storeName: storeName}
}

func (d *Dao) Put(ctx context.Context, en *Entity) error {
	bytes, err := Encode(en)
	if nil == err {
		err = d.stateClient.Set(ctx, StoreKey(en.ID), bytes)
	}
	return errors.Wrap(err, "put entity")
}

func (d *Dao) Get(ctx context.Context, id string) (en *Entity, err error) {
	var item *state.StateItem
	item, err = d.stateClient.Get(ctx, StoreKey(id))
	if nil == err {
		en = new(Entity)
		err = Decode(item.Value, en)
	}
	return en, errors.Wrap(err, "get entity")
}
