package holder

import (
	"context"
	"sync"
	"time"

	logf "github.com/tkeel-io/core/pkg/logfield"
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

func (h *holder) Cancel() {

}

func (h *holder) Wait(ctx context.Context, id string) *Waiter {
	h.lock.Lock()
	// chan size 为3 避免response和超时对chan的竞争.
	waitCh := make(chan Response, 2)
	h.holdeds[id] = waitCh
	h.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)

	go func() {
		select {
		case <-h.ctx.Done():
			log.L().Info("close holder.")
			waitCh <- Response{
				Status:  types.StatusCanceled,
				ErrCode: context.Canceled.Error(),
			}
		case <-ctx.Done():
			if nil != ctx.Err() {
				waitCh <- Response{
					Status:  types.StatusCanceled,
					ErrCode: context.Canceled.Error(),
				}
			}
		}

		// delete wait channel.
		h.lock.Lock()
		delete(h.holdeds, id)
		h.lock.Unlock()
	}()

	return &Waiter{
		ch:     waitCh,
		cancel: cancel,
	}
}

func (h *holder) OnRespond(resp *Response) {
	h.lock.Lock()
	waitCh := h.holdeds[resp.ID]
	delete(h.holdeds, resp.ID)
	h.lock.Unlock()

	log.L().Debug("received response",
		logf.ReqID(resp.ID), logf.Status(resp.Status.String()))

	if nil == waitCh {
		log.L().Warn("request terminated, user cancel or timeout", logf.ReqID(resp.ID))
		return
	}

	waitCh <- *resp
}
