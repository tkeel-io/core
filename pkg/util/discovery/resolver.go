package discovery

import (
	"context"

	"github.com/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (d *Discovery) Resolve(ctx context.Context, handlers []ResolveHandler) error {
	watchPrefix := "service://core"
	// list current nodes.
	resp, err := d.discoveryEnd.Get(ctx, watchPrefix, clientv3.WithPrefix())
	if nil != err {
		log.L().Error("list current nodes", logf.Error(err), logf.Prefix(watchPrefix))
		return errors.Wrap(err, "list current nodes")
	}

	for _, kv := range resp.Kvs {
		var node Service
		if err := json.Unmarshal(kv.Value, &node); nil != err {
			log.L().Error("unmarshal Service", logf.Error(err),
				logf.Key(string(kv.Key)), logf.Value(string(kv.Value)))
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
			log.L().Error("unmarshal Service", logf.Error(err),
				logf.Key(string(ev.Kv.Key)), logf.Value(string(ev.Kv.Value)))
		}

		// handle envent.
		for _, handler := range handlers {
			handler(EnventType(ev.Type), node)
		}
	})
	return nil
}
