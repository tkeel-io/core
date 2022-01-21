package pubsub

import "context"

type MessageHandler func(ctx context.Context, message interface{}) error

type Pubsub interface {
	Send(ctx context.Context, event interface{}) error
	Received(ctx context.Context, receiver MessageHandler) error
}
