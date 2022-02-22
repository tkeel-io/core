package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	// store mapper prefix key.
	MapperPrefix = "CORE.MAPPER"
	// CORE.MAPPER.{mapperID}.
	fmtMapperString = "%s.%s.%s"
)

type MapperHandler func([]Mapper)
type WatchMapperHandler func(EnventType, Mapper)

type Mapper struct {
	ID          string
	TQL         string
	Name        string
	EntityID    string
	EntityType  string
	Description string
}

func (m *Mapper) Copy() Mapper {
	return Mapper{
		ID:          m.ID,
		Name:        m.Name,
		TQL:         m.TQL,
		Description: m.Description,
	}
}

func (m *Mapper) Key() string {
	return fmt.Sprintf(fmtMapperString, MapperPrefix, m.EntityID, m.ID)
}

func (d *Dao) PutMapper(ctx context.Context, m *Mapper) error {
	var err error
	var bytes []byte
	if bytes, err = json.Marshal(m); nil == err {
		_, err = d.etcdEndpoint.Put(ctx, m.Key(), string(bytes))
	}
	return errors.Wrap(err, "put mapper")
}

func (d *Dao) GetMapper(ctx context.Context, m *Mapper) (*Mapper, error) {
	res, err := d.etcdEndpoint.Get(ctx, m.Key())
	if nil == err {
		if len(res.Kvs) == 0 {
			return m, errors.Wrap(xerrors.ErrMapperNotFound, "get mapper")
		}
		err = json.Unmarshal(res.Kvs[0].Value, m)
	}
	return m, errors.Wrap(err, "get mapper")
}

func (d *Dao) DelMapper(ctx context.Context, m *Mapper) error {
	_, err := d.etcdEndpoint.Delete(ctx, m.Key())
	return errors.Wrap(err, "delete mapper")
}

func (d *Dao) HasMapper(ctx context.Context, m *Mapper) (bool, error) {
	res, err := d.etcdEndpoint.Get(ctx, m.Key())
	if nil == err {
		if len(res.Kvs) == 1 {
			return true, nil
		}
		err = xerrors.ErrMapperNotFound
	}
	return false, errors.Wrap(err, "exists mapper")
}

func (d *Dao) RangeMapper(ctx context.Context, rev int64, handler MapperHandler) {
	prefix := MapperPrefix
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithRev(rev),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)))

	var count int64
	var countFailure int64
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.Error("range mapper failure", zap.Error(err), zfield.Prefix(prefix),
				zfield.Count(count), zap.Int64("failure", countFailure), zfield.Elapsedms(elapsedTime.Elapsed()))
			return
		} else if len(resp.Kvs) == 0 {
			log.Info("range mapper completed", zfield.Prefix(prefix), zfield.Count(count),
				zap.Int64("failure", countFailure), zfield.Elapsedms(elapsedTime.Elapsed()))
			return
		}

		var mappers []Mapper
		for _, kv := range resp.Kvs {
			var mapper Mapper
			if err := json.Unmarshal(kv.Value, &mapper); nil != err {
				countFailure++
				log.Error("unmarshal mapper", zap.Error(err),
					zfield.Key(string(kv.Key)), zfield.Value(string(kv.Value)))
				continue
			}
			mappers = append(mappers, mapper)
		}

		select {
		case <-ctx.Done():
			return
		default:
			handler(mappers)
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

func (d *Dao) WatchMapper(ctx context.Context, rev int64, handler WatchMapperHandler) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithPrefix(), clientv3.WithRev(rev+1))
	resp := d.etcdEndpoint.Watch(ctx, MapperPrefix, opts...)

	for {
		select {
		case <-ctx.Done():
		case wr := <-resp:
			if len(wr.Events) == 0 {
				return
			}

			for _, ev := range wr.Events {
				var mapper Mapper
				if err := json.Unmarshal(ev.Kv.Value, &mapper); nil != err {
					log.Error("unmarshal mapper", zap.Error(err),
						zfield.Key(string(ev.Kv.Key)), zfield.Value(string(ev.Kv.Value)))
					continue
				}
				handler(EnventType(ev.Type), mapper)
			}
		}
	}
}
