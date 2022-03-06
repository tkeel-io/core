package mock

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
)

func NewDispatcher() *dispatcher {
	return &dispatcher{}
}

type dispatcher struct {
}

func (d *dispatcher) Dispatch(context.Context, v1.Event) error {
	return nil
}
