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
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func (d *Dao) PutResource(ctx context.Context, res Resource) error {
	var (
		err   error
		key   []byte
		bytes []byte
	)

	if bytes, err = res.Encode(); nil != err {
		return errors.Wrap(err, "put costume resource")
	} else if key, err = res.EncodeKey(); nil == err {
		_, err = d.etcdEndpoint.Put(ctx, string(key), string(bytes))
	}

	return errors.Wrap(err, "put costume resource")
}

func (d *Dao) GetResource(ctx context.Context, res Resource) (Resource, error) {
	var (
		err error
		key []byte
		ret *clientv3.GetResponse
	)

	if key, err = res.EncodeKey(); nil != err {
		return res, errors.Wrap(err, "get costume resource")
	} else if ret, err = d.etcdEndpoint.Get(ctx, string(key)); nil == err {
		if len(ret.Kvs) == 0 {
			return res, errors.Wrap(xerrors.ErrResourceNotFound, "get costume resource")
		}
		err = res.Decode(ret.Kvs[0].Value)
	}
	return res, errors.Wrap(err, "get costume resource")
}

func (d *Dao) DelResource(ctx context.Context, res Resource) error {
	key, err := res.EncodeKey()
	if nil != err {
		return errors.Wrap(err, "delete costume resource")
	}

	_, err = d.etcdEndpoint.Delete(ctx, string(key))
	return errors.Wrap(err, "delete costume resource")
}

func (d *Dao) DelResources(ctx context.Context, prefix string) error {
	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	_, err := d.etcdEndpoint.Delete(ctx, prefix, opts...)
	return errors.Wrap(err, "delete costume resource")
}

func (d *Dao) HasResource(ctx context.Context, res Resource) (has bool, err error) {
	var key []byte
	var ret *clientv3.GetResponse
	if key, err = res.EncodeKey(); nil != err {
		return has, errors.Wrap(err, "exists costume resource")
	}

	if ret, err = d.etcdEndpoint.Get(ctx, string(key)); nil == err {
		if len(ret.Kvs) == 1 {
			return true, nil
		}
		err = xerrors.ErrResourceNotFound
	}
	return false, errors.Wrap(err, "exists costume resource")
}

func (d *Dao) ListResource(ctx context.Context, rev int64, prefix string, decodeFunc DecodeFunc) ([]Resource, error) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithPrefix())

	var count int64
	var ress []Resource
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.L().Error("list costume resource", zap.Error(err), zfield.Prefix(prefix),
				zfield.Count(count), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return ress, errors.Wrap(err, "list costume resource")
		} else if len(resp.Kvs) == 0 {
			log.L().Info("list costume resource", zfield.Prefix(prefix),
				zfield.Count(count), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return ress, nil
		}

		for _, kv := range resp.Kvs {
			var res Resource
			if res, err = decodeFunc(kv.Value); nil != err {
				log.L().Error("unmarshal costume resource", zap.Error(err),
					zfield.Key(string(kv.Key)), zfield.Value(string(kv.Value)))
				return ress, errors.Wrap(err, "unmarshal costume resource")
			}
			ress = append(ress, res)
		}

		select {
		case <-ctx.Done():
			return ress, errors.Wrap(ctx.Err(), "list costume resource")
		default:
		}

		if !resp.More {
			return ress, nil
		}
		// count all.
		count += int64(len(resp.Kvs))
		// move to next prefix.
		prefix = string(append(resp.Kvs[len(resp.Kvs)-1].Key, 0))
	}
}

func (d *Dao) RangeResource(ctx context.Context, rev int64, prefix string, handler RangeResourceFunc) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithRev(rev),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)))

	var count int64
	var countFailure int64
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.L().Error("range costume resource failure",
				zap.Error(err), zfield.Prefix(prefix),
				zfield.Elapsedms(elapsedTime.ElapsedMilli()),
				zfield.Count(count), zap.Int64("failure", countFailure))
			return
		} else if len(resp.Kvs) == 0 {
			log.L().Info("range costume resource completed",
				zap.Int64("failure", countFailure),
				zfield.Prefix(prefix), zfield.Count(count),
				zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			handler(resp.Kvs)
		}

		if !resp.More {
			return
		}
		// count all.
		count += int64(len(resp.Kvs))
		// move to next prefix.
		prefix = string(append(resp.Kvs[len(resp.Kvs)-1].Key, 0))
	}
}

func (d *Dao) WatchResource(ctx context.Context, rev int64, prefix string, handler WatchResourceFunc) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithPrefix(), clientv3.WithRev(rev+1))
	resp := d.etcdEndpoint.Watch(ctx, prefix, opts...)

	for {
		select {
		case <-ctx.Done():
		case wr := <-resp:
			if len(wr.Events) == 0 {
				return
			}

			for _, ev := range wr.Events {
				switch EnventType(ev.Type) {
				case PUT:
					handler(EnventType(ev.Type), ev.Kv)
				case DELETE:
					handler(EnventType(ev.Type), ev.Kv)
				default:
					log.L().Debug("catch event, invalid event type", zap.Any("event_type", ev.Type))
					continue
				}
			}
		}
	}
}
