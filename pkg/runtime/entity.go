package runtime

import (
	"context"

	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

type entity struct {
	id    string
	state tdtl.Collect
}

func NewEntity(id string, state []byte) (Entity, error) {
	s := tdtl.New(state)
	return &entity{id: id, state: *s}, s.Error()
}

func (e *entity) Handle(ctx context.Context, in *Result) *Result {
	var changes []Patch
	cc := e.state.Copy()
	for _, patch := range in.Patches {
		patchVal := tdtl.New(patch.Value)
		switch patch.Op {
		case OpAdd:
			cc.Append(patch.Path, patchVal)
		case OpCopy:
		case OpMerge:
			res := cc.Get(patch.Path).Merge(patchVal)
			cc.Set(patch.Path, res)
		case OpRemove:
			cc.Del(patch.Path)
		case OpReplace:
			cc.Set(patch.Path, patchVal)
		default:
			return &Result{Err: xerrors.ErrPatchPathInvalid}
		}

		if nil != cc.Error() {
			break
		}

		switch patch.Op {
		case OpMerge:
			patchVal.Foreach(func(key []byte, value *tdtl.Collect) {
				changes = append(changes, Patch{
					Op: OpReplace, Path: patch.Path, Value: value})
			})
		default:
			changes = append(changes,
				Patch{Op: patch.Op, Path: patch.Path, Value: patchVal})
		}

	}

	if cc.Error() == nil {
		e.state = *cc
	}

	return &Result{State: cc.Raw(), Changes: changes, Err: cc.Error()}
}

func (e *entity) Raw() []byte {
	return e.state.Copy().Raw()
}

func (e *entity) Basic() *tdtl.Collect {
	basic := e.state.Copy()
	basic.Del("properties", "scheme")
	return basic
}
