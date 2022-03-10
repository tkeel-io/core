package runtime

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
)

const defaultTTLMax = 5

type Patch struct {
	Op    PatchOp
	Path  string
	Value *tdtl.Collect
}

//Feed 包含实体最新状态以及变更
type Result struct {
	// TODO: 将 error 放到这里的原因： 在UpdateWithEvent无论失败还是成功，callback都是可能被执行的.
	TTL      int
	Err      error
	State    []byte
	Event    v1.Event
	EntityID string
	Patches  []Patch
	Changes  []Patch
}

type Entity interface {
	ID() string
	Handle(context.Context, *Result) *Result
	Basic() *tdtl.Collect
	Raw() []byte
}

type PersistentFunc func(interface{}) *Result
type Handler interface {
	Handle(context.Context, *Result) *Result
}

type handlerImpl struct {
	fn func(context.Context, *Result) *Result
}

func (h *handlerImpl) Handle(ctx context.Context, result *Result) *Result {
	if nil != result.Err {
		return result
	}
	return h.fn(ctx, result)
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

func NewExecer(state Entity) *Execer {
	return &Execer{
		state:    state,
		execFunc: state,
	}
}

func (e *Execer) Exec(ctx context.Context, result *Result) *Result {
	if nil != result.Err {
		return result
	}

	// execute preFuncs.
	handlers := append(e.preFuncs, e.execFunc)
	handlers = append(handlers, e.postFuncs...)

	for _, handler := range handlers {
		result = handler.Handle(ctx, result)
	}

	// 终止递归.
	if result.TTL >= defaultTTLMax {
		log.Error("ttl overflow", e.state.ID())
		return result
	} else if len(result.Patches) > 0 {
		return e.Exec(ctx, result)
	}

	result.TTL++
	return result
}

type MCache struct {
	ID        string
	Mapper    mapper.Mapper
	Tentacles []mapper.Tentacler
}
