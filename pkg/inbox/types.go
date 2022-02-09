package inbox

import (
	"context"

	"github.com/tkeel-io/core/pkg/runtime/message"
)

type MessageHandler func(msgCtx message.Context)

type Inboxer interface {
	Start()
	Consume(ctx context.Context, handler MessageHandler) error
}
