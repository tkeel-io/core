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

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type KeyValue interface {
	Close() error
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	MemberList(ctx context.Context) (*clientv3.MemberListResponse, error)
	Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error)
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
}

func newEtcd(cfg clientv3.Config) (KeyValue, error) {
	etcdEndpoint, err := clientv3.New(cfg)
	return etcdEndpoint, errors.Wrap(err, "new etcd KeyValue instance")
}

// ---------------------- KeyValue mock.

type keyValueNoop struct {
	ctx context.Context
}

func newNoop() KeyValue {
	return &keyValueNoop{}
}

func (n *keyValueNoop) Close() error { return nil }
func (n *keyValueNoop) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return &clientv3.PutResponse{}, nil
}

func (n *keyValueNoop) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return &clientv3.GetResponse{}, xerrors.ErrResourceNotFound
}

func (n *keyValueNoop) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return &clientv3.DeleteResponse{}, nil
}

func (n *keyValueNoop) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	return &clientv3.MemberListResponse{}, nil
}

func (n *keyValueNoop) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return &clientv3.StatusResponse{}, nil
}

func (n *keyValueNoop) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return make(clientv3.WatchChan)
}
