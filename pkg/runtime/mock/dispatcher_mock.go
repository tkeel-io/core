package mock

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
)

func NewDispatcher() dispatch.Dispatcher {
	return &dispatcher{}
}

type dispatcher struct{}

func (d *dispatcher) DispatchToLog(ctx context.Context, bytes []byte) error {
	panic("implement me")
}

func (d *dispatcher) Dispatch(context.Context, v1.Event) error {
	return nil
}
