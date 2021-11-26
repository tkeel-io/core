package constraint

import (
	"github.com/tkeel-io/collectjs"
)

const (
	PatchOperatorAdd     = "add"
	PatchOperatorTest    = "test" // reserved.
	PatchOperatorCopy    = "copy" // reserved.
	PatchOperatorMove    = "move" // reserved.
	PatchOperatorRemove  = "remove"
	PatchOperatorReplace = "replace" // default.
)

func Patch(destNode, srcNode Node, path, op string) (Node, error) { //nolint
	sVal := destNode.To(String)
	if String == sVal.Type() {
		s, _ := sVal.Value().(string)
		collect := collectjs.New(s)
		switch op {
		case PatchOperatorRemove:
			collect.Del(path)
		case PatchOperatorCopy:
			if collect = collect.Get(path); nil == collect {
				return nil, ErrPatchNotFound
			}
		case PatchOperatorAdd:
		case PatchOperatorReplace:
		default:
			return destNode, ErrJSONPatchReservedOp
		}

		// dispose 'remove' & 'add'
		if nil == srcNode {
			return destNode, ErrEmptyParam
		}

		switch op {
		case PatchOperatorReplace:
			collect.Set(path, ToBytesWithWrapString(srcNode))
		case PatchOperatorAdd:
			collect.Append(path, ToBytesWithWrapString(srcNode))
		}

		return JSONNode(collect.GetRaw()), nil
	}
	return destNode, ErrInvalidNodeType
}
