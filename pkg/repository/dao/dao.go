package dao

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource/state"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Dao struct {
	storeName    string
	stateClient  state.Store
	etcdEndpoint clientv3.Client
	entityCodec  entityCodec
}

func New(ctx context.Context, storeName string, stateClient state.Store) *Dao {
	return &Dao{
		storeName:   storeName,
		stateClient: stateClient,
		entityCodec: entityCodec{},
	}
}
