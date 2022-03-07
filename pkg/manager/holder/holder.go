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
	timeout time.Duration
	holdeds map[string]chan Response

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, timeout time.Duration) Holder {
	ctx, cancel := context.WithCancel(ctx)
	return &holder{
		ctx:     ctx,
		cancel:  cancel,
		lock:    sync.RWMutex{},
		timeout: timeout,
		holdeds: make(map[string]chan Response),
	}
}

func (h *holder) Wait(ctx context.Context, id string) Response {
	h.lock.Lock()
	waitCh := make(chan Response)
	h.holdeds[id] = waitCh
	h.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	var resp Response
	select {
	case <-h.ctx.Done():
		log.Info("close holder.")
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

	log.Debug("received response",
		zfield.ReqID(resp.ID), zfield.Status(resp.Status.String()))

	if nil == waitCh {
		log.Warn("request terminated, user cancel or timeout", zfield.ReqID(resp.ID))
		return
	}

	waitCh <- *resp
}
