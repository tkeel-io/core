package inbox

import (
	"context"

	"github.com/tkeel-io/core/pkg/runtime/message"
)

type MessageHandler func(ctx context.Context, msgCtx message.Context) error

type Inboxer interface {
	ID() string
	Start() error
	Close() error
	Consume(ctx context.Context, handler MessageHandler) error
}
