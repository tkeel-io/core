package dao

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/state"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Dao struct {
	stateClient  state.Store
	etcdEndpoint *clientv3.Client
	entityCodec  entityCodec
}

func New(ctx context.Context, storeCfg config.Metadata, etcdCfg config.EtcdConfig) (*Dao, error) {
	storeMeta := resource.ParseFrom(storeCfg)
	timeout := etcdCfg.DialTimeout * int64(time.Second)
	etcdEndpoint, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdCfg.Endpoints,
		DialTimeout: time.Duration(timeout),
	})

	if nil != err {
		return nil, errors.Wrap(err, "dial etcd")
	}

	return &Dao{
		etcdEndpoint: etcdEndpoint,
		entityCodec:  entityCodec{},
		stateClient:  state.NewStore(storeMeta),
	}, nil
}
