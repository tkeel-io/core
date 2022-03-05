package runtime

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
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

func (e *entity) Handle(ctx context.Context, event v1.Event) (*Result, error) {
	var changes []Patch
	ev, _ := event.(v1.PatchEvent)

	cc := e.state.Copy()
	for _, patch := range ev.Patches() {
		patchVal := tdtl.New(patch.Value)
		operator := PatchOp(patch.Operator)
		switch operator {
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
			return nil, xerrors.ErrPatchPathInvalid
		}

		if nil != cc.Error() {
			break
		}

		switch operator {
		case OpMerge:
			patchVal.Foreach(func(key []byte, value *tdtl.Collect) {
				changes = append(changes, Patch{Op: OpReplace, Path: patch.Path, Value: value})
			})
		default:
			changes = append(changes,
				Patch{Op: operator, Path: patch.Path, Value: patchVal})
		}

	}

	if cc.Error() == nil {
		e.state = *cc
	}

	return &Result{State: cc.Raw(), Changes: changes}, cc.Error()
}

func (e *entity) Raw() []byte {
	return e.state.Copy().Raw()
}
