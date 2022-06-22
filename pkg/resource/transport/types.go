package transport

import "context"

type Transport interface {
	Send(ctx context.Context, m interface{}) error
}

type Encoder func(m interface{}) interface{}
