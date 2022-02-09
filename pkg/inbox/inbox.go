package inbox

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource/pubsub"
)

// inbox 是用于实现对Queue可靠消费保证的方案实现.
// 借鉴 tcp 滑动窗口原理/

const defaultInboxCapcity int = 1000

type inbox struct {
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
		ctx:     ctx,
		cancel:  cancel,
		recever: receiver,
		capcity: defaultInboxCapcity,
	}
}

func (ix *inbox) Consume(ctx context.Context, handler MessageHandler) error {
	return nil
}

func (ix *inbox) Start() {

}
