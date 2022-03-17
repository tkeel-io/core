package runtime

import (
	"context"

	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/tdtl"
)

const defaultTTLMax = 5

type Patch struct {
	Op    PatchOp
	Path  string
	Value *tdtl.Collect
}

type EntityAttr interface {
	Type() string
	Owner() string
	Source() string
	Version() int64
	LastTime() int64
	TemplateID() string
	Properties() tdtl.Node
	Scheme() tdtl.Node
	GetProp(key string) tdtl.Node
}

type Entity interface {
	EntityAttr

	ID() string
	Get(string) tdtl.Node
	Copy() Entity
	Handle(context.Context, *Feed) *Feed
	Basic() *tdtl.Collect
	Raw() []byte
}

type handlerImpl struct {
	fn func(context.Context, *Feed) *Feed
}

func (h *handlerImpl) Handle(ctx context.Context, feed *Feed) *Feed {
	if nil != feed.Err {
		return feed
	}
	return h.fn(ctx, feed)
}

type MCache struct {
	ID        string
	EntityID  string
	Mapper    mapper.Mapper
	Tentacles []mapper.Tentacler
}

type Task func()
