package runtime2

import (
	"context"

	"github.com/pkg/errors"
)

type Inboxer interface {
	Close() error
	Consume(ctx context.Context, handler func(ContainerEvent) error) error
}

type inbox struct {
	source Source

	ctx    context.Context
	cancel context.CancelFunc
}

func NewInbox(ctx context.Context, source Source) Inboxer {
	ctx, cancel := context.WithCancel(ctx)
	return &inbox{
		ctx:    ctx,
		cancel: cancel,
		source: source,
	}
}

func (i *inbox) Close() error {
	return nil
}

func (i *inbox) Consume(ctx context.Context, handler func(ContainerEvent) error) error {
	err := i.source.StartReceiver(ctx, func(ctx context.Context, message interface{}) error {
		// 1. 将 kafka event 转换为 ConainerEvent.

		// 2. 调用 handler.
		ev := ContainerEvent{}
		handler(ev)
		return nil
	})
	return errors.Wrap(err, "consumer source")
}
