package runtime

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/tdtl"
)

type Patch struct {
	Op    PatchOp
	Path  string
	Value *tdtl.Collect
}

//Feed 包含实体最新状态以及变更
type Result struct {
	State   []byte
	Changes []Patch
}

type Entity interface {
	Handle(context.Context, v1.Event) (*Result, error)
	Raw() []byte
}

type Dispatcher interface {
	Dispatch(context.Context) error
}
