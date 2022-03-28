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
