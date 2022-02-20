/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constraint

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	xerrors "github.com/tkeel-io/core/pkg/errors"
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
	case "add": //nolint
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

func Patch(destNode, srcNode Node, path string, op PatchOperator) (Node, error) {
	bytes := ToBytesWithWrapString(destNode)
	if nil != bytes {
		collect := collectjs.ByteNew(bytes)
		switch op {
		case PatchOpRemove:
			collect.Del(path)
			return JSONNode(collect.GetRaw()), collect.GetError()
		case PatchOpCopy:
			if collect = collect.Get(path); nil == collect ||
				errors.Is(collect.GetError(), jsonparser.KeyPathNotFoundError) {
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
			setVal := ToBytesWithWrapString(srcNode)
			resBytes, err := setValue(bytes, setVal, path, op.String())
			return NewNode(resBytes), errors.Wrap(err, "patch json")
		}
		return destNode, ErrEmptyParam
	}
	return destNode, ErrInvalidNodeType
}

func check(raw []byte, path string) ([]byte, []string, []string, error) {
	// TODO: 后面可以改写 jsonparser.search 来优化.
	var (
		err       error
		valueT    jsonparser.ValueType
		prevalueT jsonparser.ValueType
	)

	segs := splitPath(path)
	if len(raw) == 0 {
		switch segs[0][0] {
		case '[':
			raw = []byte(`[]`)
		default:
			raw = []byte(`{}`)
		}
	}

	prevalueT, err = jsonparser.ParseType(raw)
	if nil != err {
		return raw, nil, nil, errors.Wrap(err, "check path")
	}

	for index := range segs {
		if _, valueT, _, err = jsonparser.Get(raw, segs[:index+1]...); nil != err {
			if segs[index][0] == byte('[') {
				if prevalueT != jsonparser.Array {
					return raw, nil, nil, xerrors.ErrInvalidJSONPath
				}
			} else if prevalueT != jsonparser.Object {
				return raw, nil, nil, xerrors.ErrInvalidJSONPath
			}

			if errors.Is(err, jsonparser.KeyPathNotFoundError) {
				return raw, segs[:index], segs[index:], nil
			}
			return raw, nil, nil, errors.Wrap(err, "check path")
		}
		prevalueT = valueT
	}

	return raw, segs, []string{}, nil
}

func setValue(raw, setVal []byte, path, op string) ([]byte, error) { //nolint
	raw2, foundSegs, notFoundSegs, err := check(raw, path)
	if nil != err {
		return raw, errors.Wrap(err, "make value")
	}

	if len(notFoundSegs) == 0 {
		switch op {
		case "add":
			if raw2, err = collectjs.Append(raw2, path, setVal); nil != err {
				return raw2, errors.Wrap(err, "set value")
			}
		case "replace":
			if raw2, err = collectjs.Set(raw2, path, setVal); nil != err {
				return raw2, errors.Wrap(err, "set value")
			}
		}
		return raw2, nil
	}

	segs := notFoundSegs[1:]
	if len(segs) > 0 {
		switch op {
		case "add":
			if setVal, err = collectjs.Append([]byte(`[]`), "", setVal); nil != err {
				return raw2, errors.Wrap(err, "set value")
			}
		}

		for index := range segs {
			seg := segs[len(segs)-1-index]
			switch seg[0] {
			case '[':
				if seg != "[0]" {
					return raw, xerrors.ErrInvalidJSONPath
				}
				if setVal, err = collectjs.Append([]byte(`[]`), "", setVal); nil != err {
					return raw2, errors.Wrap(err, "set value")
				}
			default:
				if setVal, err = collectjs.Set([]byte(`{}`), seg, setVal); nil != err {
					return raw2, errors.Wrap(err, "set value")
				}
			}
		}
	}

	path = strings.Join(append(foundSegs, notFoundSegs[0]), ".")
	path = strings.ReplaceAll(path, ".[", "[")
	if raw, err = collectjs.Set(raw2, path, setVal); nil != err {
		return raw2, errors.Wrap(err, "set value")
	}

	return raw, nil
}

func splitPath(path string) []string {
	keys := []string{}
	if len(path) > 0 {
		if path[0] == '"' && path[len(path)-1] == '"' {
			return []string{path[1 : len(path)-1]}
		}
		path = strings.ReplaceAll(path, "[", ".[")
		keys = strings.Split(path, ".")
	}
	if len(keys) > 0 && keys[0] == "" {
		return keys[1:]
	}
	return keys
}
