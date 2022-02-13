package holder

import (
	"context"
	"sync"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

type holder struct {
	holded map[string]chan Response
	lock   sync.RWMutex
}

func New() Holder {
	return &holder{
		lock:   sync.RWMutex{},
		holded: make(map[string]chan Response),
	}
}

func (h *holder) Wait(ctx context.Context, id string) Response {
	h.lock.Lock()
	waitCh := make(chan Response)
	h.holded[id] = waitCh
	h.lock.Unlock()

	var resp Response
	select {
	case <-ctx.Done():
		log.Warn("request terminated, user cancel or timeout", zfield.ID(id))
		resp = Response{
			Status:  StatusCanceled,
			ErrCode: context.Canceled.Error(),
		}
	case resp = <-waitCh:
		h.lock.Lock()
		delete(h.holded, id)
		h.lock.Unlock()
	}
	return resp
}

func (h *holder) OnRespond(resp *Response) {
	h.lock.Lock()
	waitCh := h.holded[resp.ID]
	delete(h.holded, resp.ID)
	h.lock.Unlock()

	if nil == waitCh {
		log.Warn("request terminated, user cancel or timeout", zfield.ID(resp.ID))
		return
	}

	waitCh <- *resp
}
