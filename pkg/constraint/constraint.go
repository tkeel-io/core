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
	"errors"

	"github.com/tkeel-io/tdtl"
)

const (
	EnabledFlagSelf = 1 << iota
	EnabledFlagSearch
	EnabledFlagTimeSeries
)

var (
	ErrEntityConfigInvalid = errors.New("invalid entity configurations")
	ErrJSONPatchReservedOp = errors.New("invalid json patch operator")
	ErrInvalidNodeType     = errors.New("undefine node type")
	ErrEmptyParam          = errors.New("empty params")
	ErrPatchNotFound       = errors.New("patch not found")
	ErrPatchPathInvalid    = errors.New("invalid patch path")
	ErrPatchPathLack       = errors.New("patch path lack")
	ErrPatchPathRoot       = errors.New("patch path lack root")
	ErrPatchTypeInvalid    = errors.New("patch config type invalid")
)

var callbacks = map[string]func(op Operator, val tdtl.Node) (tdtl.Node, error){
	"max": func(op Operator, val tdtl.Node) (tdtl.Node, error) {
		return val, nil
	},
	"size": func(op Operator, val tdtl.Node) (tdtl.Node, error) {
		return val, nil
	},
}

type Operator struct {
	Callback  string
	Condition interface{}
}

func newOperator(cb string, cond interface{}) Operator {
	return Operator{
		Callback:  cb,
		Condition: cond,
	}
}

type Constraint struct {
	ID         string
	Type       string
	Operators  []Operator
	ChildNodes []*Constraint
	EnableFlag *BitBucket
}

func newConstraint() *Constraint {
	return &Constraint{EnableFlag: NewBitBucket(8)}
}

func (ct *Constraint) GenEnabledIndexes(enabledFlag int) []string {
	return genEnabledIndexes("", enabledFlag, ct)
}

func genEnabledIndexes(prefix string, enabledFlag int, ct *Constraint) []string {
	var searchIndexes []string
	if !ct.EnableFlag.Enabled(EnabledFlagSelf) {
		return []string{}
	}

	if ct.EnableFlag.Enabled(enabledFlag) {
		searchIndexes = append(searchIndexes, prefix+ct.ID)
	}

	for _, childCt := range ct.ChildNodes {
		searchIndexes = append(searchIndexes, genEnabledIndexes(ct.ID+".", enabledFlag, childCt)...)
	}
	return searchIndexes
}

func NewConstraintsFrom(cfg Config) *Constraint {
	return parseConstraintFrom(cfg)
}

func parseConstraintFrom(cfg Config) *Constraint {
	// current latyer.
	if !cfg.Enabled {
		return nil
	}

	ct := newConstraint()

	ct.ID = cfg.ID
	ct.Type = cfg.Type
	ct.EnableFlag.Enable(EnabledFlagSelf)
	if cfg.EnabledSearch {
		ct.EnableFlag.Enable(EnabledFlagSearch)
	}
	if cfg.EnabledTimeSeries {
		ct.EnableFlag.Enable(EnabledFlagTimeSeries)
	}

	switch cfg.Type {
	case PropertyTypeArray:
		define := cfg.getArrayDefine()
		if childCt := parseConstraintFrom(define.ElemType); nil != childCt {
			ct.ChildNodes = append(ct.ChildNodes, childCt)
		}
	case PropertyTypeStruct:
		define := cfg.getStructDefine()
		for _, field := range define.Fields {
			if childCt := parseConstraintFrom(field); nil != childCt {
				ct.ChildNodes = append(ct.ChildNodes, childCt)
			}
		}
	default:
		// TODO: .
	}

	// parse define.
	ct.Operators = parseDefine(cfg.Define)

	return ct
}

func parseDefine(define map[string]interface{}) []Operator {
	var ops []Operator
	for key, val := range define {
		if keyContains(key) {
			ops = append(ops, newOperator(key, val))
		}
	}
	return ops
}

func keyContains(key string) bool {
	_, flag := callbacks[key]
	return flag
}

func ExecData(val tdtl.Node, ct *Constraint) (tdtl.Node, error) {
	return val, nil
}
