package discovery

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func (d *Discovery) Resolve(ctx context.Context, handlers []ResolveHandler) error {
	watchPrefix := "service://core"
	// list current nodes.
	resp, err := d.discoveryEnd.Get(ctx, watchPrefix, clientv3.WithPrefix())
	if nil != err {
		log.Error("list current nodes", zap.Error(err), zfield.Prefix(watchPrefix))
		return errors.Wrap(err, "list current nodes")
	}

	for _, kv := range resp.Kvs {
		var node Service
		if err := json.Unmarshal(kv.Value, &node); nil != err {
			log.Error("unmarshal Service", zap.Error(err),
				zfield.Key(string(kv.Key)), zfield.Value(string(kv.Value)))
			continue
		}
		// handle envent.
		for _, handler := range handlers {
			handler(PUT, node)
		}
	}

	watcher := NewWatcherWithClient(ctx, d.discoveryEnd)
	watcher.Watch(watchPrefix, true, func(ev *clientv3.Event) {
		var node Service
		if err := json.Unmarshal(ev.Kv.Value, &node); nil != err {
			log.Error("unmarshal Service", zap.Error(err),
				zfield.Key(string(ev.Kv.Key)), zfield.Value(string(ev.Kv.Value)))
		}

		// handle envent.
		for _, handler := range handlers {
			handler(EnventType(ev.Type), node)
		}
	})
	return nil
}
