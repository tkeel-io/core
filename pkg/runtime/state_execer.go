package runtime

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

/*

实现state的处理循环.

*/

type Handler interface {
	Handle(context.Context, *Feed) *Feed
}

type Feed struct {
	TTL      int
	Err      error
	Event    v1.Event
	State    []byte
	EntityID string
	Patches  []Patch
	Changes  []Patch
}

// The *Funcs functions are executed in the following order:
//   * preFuncs()
//   * execFunc()
//   * postFuncs()
type Execer struct {
	state     Entity
	preFuncs  []Handler
	execFunc  Handler
	postFuncs []Handler
}

func (e *Execer) Exec(ctx context.Context, feed *Feed) *Feed {
	if nil != feed.Err {
		return feed
	}

	// handle preFuncs.
	for _, handler := range e.preFuncs {
		feed = handler.Handle(ctx, feed)
	}

	// handle execFunc.
	feed = e.execFunc.Handle(ctx, feed)

	// handle postFuncs.
	for _, handler := range e.postFuncs {
		feed = handler.Handle(ctx, feed)
	}

	if feed.TTL >= defaultTTLMax {
		log.L().Error("ttl overflow", zfield.Eid(e.state.ID()))
		return feed
	}

	feed.TTL++
	return feed
}
