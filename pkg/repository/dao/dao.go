package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/config"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/store"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type Dao struct {
	storeCfg     config.Metadata
	etcdCfg      config.EtcdConfig
	stateClient  store.Store
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
		etcdCfg:      etcdCfg,
		storeCfg:     storeCfg,
		etcdEndpoint: etcdEndpoint,
		entityCodec:  entityCodec{},
		stateClient:  store.NewStore(storeMeta),
	}, nil
}

func (d *Dao) GetLastRevision(ctx context.Context) int64 {
	var err error
	var res *clientv3.MemberListResponse
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if res, err = d.etcdEndpoint.MemberList(ctx); err != nil {
		log.Error("query etcd cluster member", zap.Error(err))
	}

	rev := int64(0)
	for _, node := range res.Members {
		log.Info("etcd node information", zfield.Name(node.Name),
			zfield.ID(fmt.Sprintf("%v", node.ID)), zap.Any("URL", node.ClientURLs))
		for _, url := range node.ClientURLs {
			resp, err := d.etcdEndpoint.Status(ctx, url)
			if err != nil {
				log.Warn("query etcd node status", zfield.Name(node.Name),
					zap.Error(err), zfield.ID(fmt.Sprintf("%v", node.ID)), zap.Any("URL", url))
				continue
			}
			if resp.Header.Revision == 0 {
				log.Fatal("zero revision")
			}
			if rev == 0 || rev > resp.Header.Revision {
				rev = resp.Header.Revision
			}
		}
	}
	return rev
}
