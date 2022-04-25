/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	ctx          context.Context
	cancel       context.CancelFunc
	storeCfg     config.Metadata
	etcdCfg      config.EtcdConfig
	stateClient  store.Store
	etcdEndpoint KeyValue
}

func NewMock(ctx context.Context, storeCfg config.Metadata, etcdCfg config.EtcdConfig) (IDao, error) {
	storeMeta := resource.ParseFrom(storeCfg)

	// create Dao instance.
	ctx, cancel := context.WithCancel(ctx)
	return &Dao{
		ctx:          ctx,
		cancel:       cancel,
		etcdCfg:      etcdCfg,
		storeCfg:     storeCfg,
		etcdEndpoint: newNoop(),
		stateClient:  store.NewStore(storeMeta),
	}, nil
}

func New(ctx context.Context, storeCfg config.Metadata, etcdCfg config.EtcdConfig) (IDao, error) {
	storeMeta := resource.ParseFrom(storeCfg)
	timeout := etcdCfg.DialTimeout * int64(time.Second)
	etcdEndpoint, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdCfg.Endpoints,
		DialTimeout: time.Duration(timeout),
	})

	if nil != err {
		return nil, errors.Wrap(err, "dial etcd")
	}

	// create Dao instance.
	ctx, cancel := context.WithCancel(ctx)
	return &Dao{
		ctx:          ctx,
		cancel:       cancel,
		etcdCfg:      etcdCfg,
		storeCfg:     storeCfg,
		etcdEndpoint: etcdEndpoint,
		stateClient:  store.NewStore(storeMeta),
	}, nil
}

func (d *Dao) GetLastRevision(ctx context.Context) int64 {
	var err error
	var res *clientv3.MemberListResponse
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if res, err = d.etcdEndpoint.MemberList(ctx); err != nil {
		log.L().Error("query etcd cluster member", zap.Error(err))
		return 0
	}

	rev := int64(0)
	for _, node := range res.Members {
		log.L().Info("etcd node information", zfield.Name(node.Name),
			zfield.ID(fmt.Sprintf("%v", node.ID)), zap.Any("URL", node.ClientURLs))
		for _, url := range node.ClientURLs {
			resp, err := d.etcdEndpoint.Status(ctx, url)
			if err != nil {
				log.L().Warn("query etcd node status", zfield.Name(node.Name),
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

func (d *Dao) Close() {
	d.cancel()
	d.etcdEndpoint.Close()
}
