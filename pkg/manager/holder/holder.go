package holder

import (
	"context"
	"sync"
	"time"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/kit/log"
)

type holder struct {
	holdeds map[string]chan Response
	lock    sync.RWMutex
}

func New() Holder {
	return &holder{
		lock:    sync.RWMutex{},
		holdeds: make(map[string]chan Response),
	}
}

func (h *holder) Wait(ctx context.Context, id string) Response {
	h.lock.Lock()
	waitCh := make(chan Response)
	h.holdeds[id] = waitCh
	h.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var resp Response
	select {
	case <-ctx.Done():
		log.Warn("request terminated, user cancel or timeout", zfield.ID(id))
		resp = Response{
			Status:  types.StatusCanceled,
			ErrCode: context.Canceled.Error(),
		}
	case resp = <-waitCh:
		log.Debug("core.API call completed", zfield.ReqID(id))
	}

	// delete waiter.
	h.lock.Lock()
	delete(h.holdeds, id)
	h.lock.Unlock()

	return resp
}

func (h *holder) OnRespond(resp *Response) {
	h.lock.Lock()
	waitCh := h.holdeds[resp.ID]
	delete(h.holdeds, resp.ID)
	h.lock.Unlock()

	if nil == waitCh {
		log.Warn("request terminated, user cancel or timeout", zfield.ID(resp.ID))
		return
	}

	waitCh <- *resp
}
