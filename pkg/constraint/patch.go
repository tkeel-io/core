package constraint

import (
	"strings"

	"github.com/tkeel-io/collectjs"
)

type PatchOperator int

// reference: https://datatracker.ietf.org/doc/html/rfc6902 .
// implement [ add, remove, replace ], reversed [ copy, move, test ].
const (
	PatchOpUndef PatchOperator = iota
	PatchOpAdd
	PatchOpTest
	PatchOpCopy
	PatchOpMove
	PatchOpRemove
	PatchOpReplace
)

func NewPatchOperator(op string) PatchOperator {
	switch op {
	case "add":
		return PatchOpAdd
	case "move":
		return PatchOpMove
	case "copy":
		return PatchOpCopy
	case "test":
		return PatchOpTest
	case "remove":
		return PatchOpRemove
	case "replace":
		return PatchOpReplace
	default:
		return PatchOpReplace
	}
}

func (po PatchOperator) String() string {
	switch po {
	case PatchOpAdd:
		return "add"
	case PatchOpMove:
		return "move"
	case PatchOpCopy:
		return "copy"
	case PatchOpTest:
		return "test"
	case PatchOpRemove:
		return "remove"
	case PatchOpReplace:
		return "replace"
	default:
		return "undefine"
	}
}

func IsReversedOp(op string) bool {
	switch op {
	case "add", "remove", "replace":
		return false
	default:
		return true
	}
}

func IsValidPath(path string) bool {
	if path == "" || strings.HasPrefix(path, ".") || strings.HasSuffix(path, ".") {
		return false
	}
	return true
}

func Patch(destNode, srcNode Node, path string, op PatchOperator) (Node, error) { //nolint
	bytes := ToBytesWithWrapString(destNode)
	if nil != bytes {
		collect := collectjs.ByteNew(bytes)
		switch op {
		case PatchOpRemove:
			collect.Del(path)
			return JSONNode(collect.GetRaw()), collect.GetError()
		case PatchOpCopy:
			if collect = collect.Get(path); nil == collect {
				return nil, ErrPatchNotFound
			}
			return JSONNode(collect.GetRaw()), collect.GetError()
		case PatchOpAdd:
		case PatchOpReplace:
		default:
			return destNode, ErrJSONPatchReservedOp
		}

		// dispose 'remove' & 'add'
		if nil != srcNode {
			switch op {
			case PatchOpReplace:
				collect.Set(path, ToBytesWithWrapString(srcNode))
			case PatchOpAdd:
				collect.Append(path, ToBytesWithWrapString(srcNode))
			}
			return JSONNode(collect.GetRaw()), collect.GetError()
		}

		return destNode, ErrEmptyParam
	}
	return destNode, ErrInvalidNodeType
}
