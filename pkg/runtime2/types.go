package runtime3

import "context"

type Patch struct {
	Op    string
	Path  string
	Value []byte
}

//Feed 包含实体最新状态以及变更
type Result struct {
	State   []byte
	Changes []Patch
}

type Entity interface {
	Handle(ctx context.Context, message interface{}) (*Result, error)
	Raw() ([]byte, error)
}

type Dispatcher interface {
	Dispatch(context.Context) error
}
