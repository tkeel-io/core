package dao

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource/state"
)

type Dao struct {
	storeName   string
	stateClient state.Store
}

func New(ctx context.Context, storeName string, stateClient state.Store) *Dao {
	return &Dao{stateClient: stateClient, storeName: storeName}
}
