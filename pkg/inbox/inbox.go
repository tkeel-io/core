package inbox

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// inbox 是用于实现对Queue可靠消费保证的方案实现.
// 借鉴 tcp 滑动窗口原理/

const defaultInboxCapcity int = 1000

type IDKey struct{}

type inbox struct {
	id           string
	size         int
	capcity      int
	commitedIdx  int
	deliveredIdx int
	recever      pubsub.Receiver
	ctx          context.Context
	cancel       context.CancelFunc
}

func New(ctx context.Context, receiver pubsub.Receiver) Inboxer {
	ctx, cancel := context.WithCancel(ctx)
	return &inbox{
		id:      receiver.ID(),
		ctx:     ctx,
		cancel:  cancel,
		recever: receiver,
		capcity: defaultInboxCapcity,
	}
}

func (ix *inbox) ID() string {
	return ix.id
}

func (ix *inbox) Consume(ctx context.Context, handler MessageHandler) error {
	err := ix.recever.Received(ctx, func(ctx context.Context, ev cloudevents.Event) error {
		msgCtx, err := message.From(ctx, ev)
		if nil != err {
			log.Error("parse event", zap.Error(err))
			return errors.Wrap(err, "consume inbox")
		}

		ctx = context.WithValue(ctx, IDKey{}, ix.id)

		handler(ctx, msgCtx)
		return nil
	})
	return errors.Wrap(err, "consume inbox")
}

func (ix *inbox) Start() error {
	return nil
}

func (ix *inbox) Close() error {
	ix.size = 0
	ix.capcity = 0
	ix.commitedIdx = 0
	ix.deliveredIdx = 0

	ix.cancel()
	err := ix.recever.Close()
	log.Info("close inbox", zfield.ID(ix.id))
	return errors.Wrap(err, "close inbox")
}
