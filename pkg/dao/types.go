package dao

import (
	"context"
	"errors"
)

const EntityStorePrefix = "core.entity."

var ErrEntityInvalidProps = errors.New("invalid entity properties")

type IDao interface {
	Get(ctx context.Context, id string) (en *Entity, err error)
	Put(ctx context.Context, en *Entity) error
}

func StoreKey(id string) string {
	return EntityStorePrefix + id
}
