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

package discovery

import (
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

type Watcher interface {
	Watch(string, bool, func(*clientv3.Event))
	Shutdown()
}

type watcher struct {
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func (w *watcher) Close() {
	w.cancel()
}

func NewWatcher(ctx context.Context, brokers []string) (Watcher, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   brokers,
		DialTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithCancel(ctx)
	return &watcher{
		client: cli,
		ctx:    ctx,
		cancel: cancel,
	}, errors.Wrap(err, "create watcher failed")
}

func NewWatcherWithClient(ctx context.Context, cli *clientv3.Client) Watcher {
	ctx, cancel := context.WithCancel(ctx)
	return &watcher{client: cli, ctx: ctx, cancel: cancel}
}

func (w *watcher) Watch(key string, prefix bool, handler func(*clientv3.Event)) {
	go func() {
		opts := []clientv3.OpOption{}
		if prefix {
			opts = append(opts, clientv3.WithPrefix())
		}

		rch := w.client.Watch(w.ctx, key, opts...)
		for {
			select {
			case wresp, ok := <-rch:
				if !ok {
					log.L().Info("channel closed, watcher exit.")
					return
				}

				for _, ev := range wresp.Events {
					handler(ev)
				}
			case <-w.ctx.Done():
				log.L().Info("watcher exot")
				return
			}
		}
	}()
}

func (w *watcher) Shutdown() {
	w.cancel()
}
