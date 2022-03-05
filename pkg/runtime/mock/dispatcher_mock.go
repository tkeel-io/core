package mock

import "context"

func NewDispatcher() *dispatcher {
	return &dispatcher{}
}

type dispatcher struct {
}

func (d *dispatcher) Dispatch(context.Context) error {
	return nil
}
