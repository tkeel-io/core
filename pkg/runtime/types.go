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
	// TODO: 将 error 放到这里的原因： 在UpdateWithEvent无论失败还是成功，callback都是可能被执行的.
	Err     error
	State   []byte
	Changes []Patch
}

type Entity interface {
	Handle(context.Context, v1.Event) *Result
	Raw() []byte
}
