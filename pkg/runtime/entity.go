package runtime

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

type entity struct {
	id    string
	state tdtl.Collect
}

func DefaultEntity(id string) Entity {
	return &entity{id: id, state: *tdtl.New([]byte(`{}`))}
}

func NewEntity(id string, state []byte) (Entity, error) {
	s := tdtl.New(state)
	s.Set("scheme", tdtl.New([]byte("{}")))
	return &entity{id: id, state: *s}, errors.Wrap(s.Error(), "new entity")
}

func (e *entity) ID() string {
	return e.id
}

func (e *entity) Get(path string) tdtl.Node {
	return e.state.Get(path)
}

func (e *entity) Handle(ctx context.Context, feed *Feed) *Feed {
	if nil != feed.Err {
		return feed
	}

	var changes []Patch
	cc := e.state.Copy()
	for _, patch := range feed.Patches {
		switch patch.Op {
		case OpAdd:
			cc.Append(patch.Path, patch.Value)
		case OpCopy:
		case OpMerge:
			res := cc.Get(patch.Path).Merge(patch.Value)
			cc.Set(patch.Path, res)
		case OpRemove:
			cc.Del(patch.Path)
		case OpReplace:
			cc.Set(patch.Path, patch.Value)
		default:
			return &Feed{Err: xerrors.ErrPatchPathInvalid}
		}

		if nil != cc.Error() {
			break
		}

		switch patch.Op {
		case OpMerge:
			patch.Value.Foreach(func(key []byte, value *tdtl.Collect) {
				changes = append(changes, Patch{
					Op: OpReplace, Value: value,
					Path: strings.Join([]string{patch.Path, string(key)}, ".")})
			})
		default:
			changes = append(changes,
				Patch{Op: patch.Op, Path: patch.Path, Value: patch.Value})
		}
	}

	if cc.Error() == nil {
		e.state = *cc
	}

	// in.Patches 处理完毕，丢弃.
	feed.Err = cc.Error()
	feed.Changes = changes
	feed.Patches = []Patch{}
	feed.State = e.Raw()
	return feed
}

func (e *entity) Raw() []byte {
	return e.state.Copy().Raw()
}

func (e *entity) Copy() Entity {
	cp := e.state.Copy()
	return &entity{
		id:    e.id,
		state: *cp,
	}
}

func (e *entity) Basic() *tdtl.Collect {
	basic := e.state.Copy()
	basic.Set("scheme", tdtl.New([]byte("{}")))
	basic.Set("properties", tdtl.New([]byte("{}")))
	return basic
}