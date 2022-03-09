package holder

import (
	"context"

	"github.com/tkeel-io/core/pkg/types"
)

type Holder interface {
	Cancel()
	Wait(ctx context.Context, id string) *Waiter
	OnRespond(*Response)
}

type Response struct {
	ID       string
	Status   types.Status
	ErrCode  string
	Metadata map[string]string
	Data     []byte
}

type Waiter struct {
	ch     chan Response
	cancel context.CancelFunc
}

func (w *Waiter) Wait() Response {
	resp := <-w.ch
	return resp
}

func (w *Waiter) Cancel() {
	w.cancel()
}
