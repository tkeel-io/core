package dao

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ListCostumeResourceFunc func([]*mvccpb.KeyValue)
type WatchCostumeResourceFunc func(EnventType, *mvccpb.KeyValue)

type CostumeResource interface {
	Key() string
	EKey() string
}

type CostumeResourceReq interface {
	Owner()    string
	EntityID() string
	Prefix() string
}

func (d *Dao) PutCostumeResource(ctx context.Context, expr CostumeResource) error {
	var err error
	var bytes []byte
	if bytes, err = json.Marshal(expr); nil == err {
		_, err = d.etcdEndpoint.Put(ctx, expr.Key(), string(bytes))
	}
	return errors.Wrap(err, "put expression")
}

func (d *Dao) GetCostumeResource(ctx context.Context, expr CostumeResource) (CostumeResource, error) {
	res, err := d.etcdEndpoint.Get(ctx, expr.Key())
	if nil == err {
		if len(res.Kvs) == 0 {
			return expr, errors.Wrap(xerrors.ErrMapperNotFound, "get expression")
		}
		err = json.Unmarshal(res.Kvs[0].Value, &expr)
	}
	return expr, errors.Wrap(err, "get expression")
}

func (d *Dao) DelCostumeResource(ctx context.Context, expr CostumeResource) error {
	_, err := d.etcdEndpoint.Delete(ctx, expr.Key())
	return errors.Wrap(err, "delete expression")
}

//@TODO
//func (d *Dao) DelByEnity(ctx context.Context, expr CostumeResource) error {
//	_, err := d.etcdEndpoint.Delete(ctx, expr.Key(), clientv3.WithPrefix())
//	return errors.Wrap(err, "delete expression by entity")
//}

func (d *Dao) HasCostumeResource(ctx context.Context, expr CostumeResource) (bool, error) {
	res, err := d.etcdEndpoint.Get(ctx, expr.Key())
	if nil == err {
		if len(res.Kvs) == 1 {
			return true, nil
		}
		err = xerrors.ErrMapperNotFound
	}
	return false, errors.Wrap(err, "exists expression")
}

func (d *Dao) ListCostumeResource(ctx context.Context, rev int64, req CostumeResourceReq) ([]CostumeResource, error) {
	// construct mapper prefix key.
	arr := []string{fmtExprPrefix, req.Owner()}
	if req.Owner() == "" {
		return nil, xerrors.ErrEmptyParam
	} else if req.EntityID() != "" {
		arr = append(arr, req.EntityID())
	}

	prefix := req.Prefix()
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithRev(rev),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)))

	var count int64
	var exprs []CostumeResource
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.L().Error("list expression", zap.Error(err), zfield.Prefix(prefix),
				zfield.Count(count), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return exprs, errors.Wrap(err, "list expression")
		} else if len(resp.Kvs) == 0 {
			log.L().Info("list expression", zfield.Prefix(prefix),
				zfield.Count(count), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return exprs, nil
		}

		for _, kv := range resp.Kvs {
			var expr CostumeResource
			if err = json.Unmarshal(kv.Value, &expr); nil != err {
				log.L().Error("unmarshal expression", zap.Error(err),
					zfield.Key(string(kv.Key)), zfield.Value(string(kv.Value)))
				return exprs, errors.Wrap(err, "unmarshal expression")
			}
			exprs = append(exprs, expr)
		}

		select {
		case <-ctx.Done():
			return exprs, errors.Wrap(ctx.Err(), "list expression")
		default:
		}

		if !resp.More {
			return exprs, nil
		}
		// count all.
		count += int64(len(resp.Kvs))
		// move to next prefix.
		prefix = string(append(resp.Kvs[len(resp.Kvs)-1].Key, 0))
	}
}

func (d *Dao) RangeCostumeResource(ctx context.Context, rev int64, handler ListCostumeResourceFunc) {
	prefix := fmtExprPrefix
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithRev(rev),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)))

	var count int64
	var countFailure int64
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.L().Error("range expression failure",
				zap.Error(err), zfield.Prefix(prefix),
				zfield.Elapsedms(elapsedTime.ElapsedMilli()),
				zfield.Count(count), zap.Int64("failure", countFailure))
			return
		} else if len(resp.Kvs) == 0 {
			log.L().Info("range expression completed",
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

func (d *Dao) WatchCostumeResource(ctx context.Context, rev int64, handler WatchCostumeResourceFunc) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithPrefix(), clientv3.WithRev(rev+1))
	resp := d.etcdEndpoint.Watch(ctx, fmtExprPrefix, opts...)

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
