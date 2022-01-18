package dao

import (
	"errors"
)

const EntityStorePrefix = "core.entity."

var ErrEntityInvalidProps = errors.New("invalid entity properties")

func StoreKey(id string) string {
	return EntityStorePrefix + id
}
