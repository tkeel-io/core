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
	// store queue prefix key.
	QueuePrefix = "CORE.QUEUE"
	// CORE.QUEUE.{id} .
	fmtQueueString = "%s.%s"
)

type QueueHandler func([]Queue)
type WatchQueueHandler func(EnventType, Queue)

type QueueType string
type ConsumerType string

func (qt QueueType) String() string    { return string(qt) }
func (ct ConsumerType) String() string { return string(ct) }

const (
	// enumerate queue type.
	QueueTypeDapr           QueueType = "dapr"
	QueueTypeKafkaTopic     QueueType = "kafka.topic"
	QueueTypeKafkaPartition QueueType = "kafka.topic.partition"

	// enumerate consumer type.
	ConsumerTypeCore     ConsumerType = "core"
	ConsumerTypeDispatch ConsumerType = "dispatcher"
)

type Queue struct {
	ID           string
	Name         string
	Type         QueueType
	Version      int64
	NodeName     string   // NodeName for core.
	Consumers    []string // Consumers for dispatcher.
	ConsumerType ConsumerType
	Description  string
	Metadata     map[string]interface{}
}

func (q *Queue) Check() error {
	// check queue type.
	switch q.Type {
	case QueueTypeKafkaTopic:
	case QueueTypeKafkaPartition:
	default:
		return xerrors.ErrInvalidQueueType
	}

	// check consumer type.
	switch q.ConsumerType {
	case ConsumerTypeCore:
	case ConsumerTypeDispatch:
	default:
		return xerrors.ErrInvalidQueueConsumerType
	}

	return nil
}

func (q *Queue) Key() string {
	return fmt.Sprintf(fmtQueueString, QueuePrefix, q.ID)
}

func (d *Dao) PutQueue(ctx context.Context, q *Queue) error {
	var err error
	var bytes []byte
	if bytes, err = json.Marshal(q); nil == err {
		_, err = d.etcdEndpoint.Put(ctx, q.Key(), string(bytes))
	}
	return errors.Wrap(err, "put queue")
}

func (d *Dao) GetQueue(ctx context.Context, q *Queue) (*Queue, error) {
	res, err := d.etcdEndpoint.Get(ctx, q.Key())
	if nil == err {
		if len(res.Kvs) == 0 {
			return q, xerrors.ErrQueueNotFound
		}
		err = json.Unmarshal(res.Kvs[0].Value, q)
	}
	return q, errors.Wrap(err, "get queue")
}

func (d *Dao) DelQueue(ctx context.Context, q *Queue) error {
	_, err := d.etcdEndpoint.Delete(ctx, q.Key())
	return errors.Wrap(err, "delete queue")
}

func (d *Dao) HasQueue(ctx context.Context, q *Queue) (bool, error) {
	res, err := d.etcdEndpoint.Get(ctx, q.Key())
	if nil == err {
		if len(res.Kvs) == 1 {
			return true, nil
		}
		err = xerrors.ErrQueueNotFound
	}
	return false, errors.Wrap(err, "exists queue")
}

func (d *Dao) RangeQueue(ctx context.Context, rev int64, handler QueueHandler) {
	prefix := QueuePrefix
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithRev(rev),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)))

	var count int64
	var countFailure int64
	var elapsedTime = util.NewElapsed()
	for {
		resp, err := d.etcdEndpoint.Get(ctx, prefix, opts...)
		if err != nil {
			log.L().Error("range queue failure", zap.Error(err), zfield.Prefix(prefix),
				zfield.Count(count), zap.Int64("failure", countFailure), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return
		} else if len(resp.Kvs) == 0 {
			log.L().Info("range queue completed", zfield.Prefix(prefix), zfield.Count(count),
				zap.Int64("failure", countFailure), zfield.Elapsedms(elapsedTime.ElapsedMilli()))
			return
		}

		var queues []Queue
		for _, kv := range resp.Kvs {
			var queue Queue
			if err := json.Unmarshal(kv.Value, &queue); nil != err {
				countFailure++
				log.L().Error("unmarshal queue", zap.Error(err),
					zfield.Key(string(kv.Key)), zfield.Value(string(kv.Value)))
				continue
			}
			queues = append(queues, queue)
		}

		select {
		case <-ctx.Done():
			return
		default:
			handler(queues)
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

func (d *Dao) WatchQueue(ctx context.Context, rev int64, handler WatchQueueHandler) {
	opts := make([]clientv3.OpOption, 0)
	opts = append(opts, clientv3.WithPrefix(), clientv3.WithRev(rev+1))
	resp := d.etcdEndpoint.Watch(ctx, QueuePrefix, opts...)

	for {
		select {
		case <-ctx.Done():
		case wr := <-resp:
			if len(wr.Events) == 0 {
				return
			}

			for _, ev := range wr.Events {
				var queue Queue
				if err := json.Unmarshal(ev.Kv.Value, &queue); nil != err {
					log.L().Error("unmarshal queue", zap.Error(err),
						zfield.Key(string(ev.Kv.Key)), zfield.Value(string(ev.Kv.Value)))
					continue
				}

				handler(EnventType(ev.Type), queue)
			}
		}
	}
}
