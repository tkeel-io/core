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

package json

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

type PatchOp int

// reference: https://datatracker.ietf.org/doc/html/rfc6902 .
// implement [ add, remove, replace ], reversed [ copy, move, test ].
const (
	OpUndef PatchOp = iota
	OpAdd
	OpTest
	OpCopy
	OpMove
	OpMerge
	OpRemove
	OpReplace
)

func NewPatchOp(op string) PatchOp {
	switch op {
	case "add":
		return OpAdd
	case "move":
		return OpMove
	case "copy":
		return OpCopy
	case "test":
		return OpTest
	case "merge":
		return OpMerge
	case "remove":
		return OpRemove
	case "replace":
		return OpReplace
	default:
		return OpUndef
	}
}

func (po PatchOp) String() string {
	switch po {
	case OpAdd:
		return "add"
	case OpMove:
		return "move"
	case OpCopy:
		return "copy"
	case OpTest:
		return "test"
	case OpMerge:
		return "merge"
	case OpRemove:
		return "remove"
	case OpReplace:
		return "replace"
	default:
		return "undefine"
	}
}

func IsReversedOp(op string) bool {
	switch op {
	case "add", "remove", "replace", "copy":
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

func Patch(destNode, srcNode tdtl.Node, path string, op PatchOp) (tdtl.Node, error) {
	var (
		err     error
		bytes   []byte
		rawData = destNode.Raw()
	)

	switch op {
	case OpRemove:
		bytes = collectjs.Del(rawData, path)
		return tdtl.New(bytes), errors.Wrap(err, "patch remove")
	case OpCopy:
		bytes, _, err = collectjs.Get(rawData, path)
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return tdtl.NULL_RESULT, xerrors.ErrPropertyNotFound
		}
		return tdtl.New(bytes), errors.Wrap(err, "patch copy")
	case OpAdd:
	case OpReplace:
	default:
		return destNode, xerrors.ErrJSONPatchReservedOp
	}

	// dispose 'remove' & 'add'.
	if nil != srcNode {
		resBytes, err := setValue(rawData, srcNode.Raw(), path, op)
		return tdtl.New(resBytes), errors.Wrap(err, "patch json")
	}
	return destNode, xerrors.ErrEmptyParam
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
		return raw, nil, nil, errors.Wrap(err, "check path 1")
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
			return raw, nil, nil, errors.Wrap(err, "check path 2")
		}
		prevalueT = valueT
	}

	return raw, segs, []string{}, nil
}

func setValue(raw, setVal []byte, path string, op PatchOp) ([]byte, error) { //nolint
	raw2, foundSegs, notFoundSegs, err := check(raw, path)
	if nil != err {
		return raw, errors.Wrap(err, "make value")
	}

	if len(notFoundSegs) == 0 {
		switch op {
		case OpAdd:
			if raw2, err = collectjs.Append(raw2, path, setVal); nil != err {
				return raw2, errors.Wrap(err, "set value")
			}
		case OpReplace:
			if raw2, err = collectjs.Set(raw2, path, setVal); nil != err {
				return raw2, errors.Wrap(err, "set value")
			}
		}
		return raw2, nil
	}

	switch op {
	case OpAdd:
		if setVal, err = collectjs.Append([]byte(`[]`), "", setVal); nil != err {
			return raw2, errors.Wrap(err, "set value")
		}
	}

	segs := notFoundSegs[1:]
	if len(segs) > 0 {
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
