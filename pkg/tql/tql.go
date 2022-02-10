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

package tql

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
)

type tql struct {
	text     string
	config   TQLConfig
	listener Listener
}

func NewTQL(tqlString string) (TQL, error) {
	var cfg TQLConfig
	listener, err := Parse(tqlString)
	if nil != err {
		return nil, errors.Wrap(err, "parse TQL")
	} else if cfg, err = listener.GetParseConfigs(); nil != err {
		return nil, errors.Wrap(err, "parse TQL and get configurations")
	}
	return &tql{text: tqlString, listener: listener, config: cfg}, nil
}

// Target returns target entity.
func (t *tql) Target() string {
	return t.config.TargetEntity
}

// Entities returns source entities.
func (t *tql) Entities() []string {
	return t.config.SourceEntities
}

// Tentacles returns tentacles.
func (t *tql) Tentacles() []TentacleConfig {
	return t.config.Tentacles
}

// Exec execute MQL.
func (t *tql) Exec(in map[string]constraint.Node) (map[string]constraint.Node, error) {
	input := make(map[string][]byte)
	for key, val := range in {
		switch val.Type() {
		case constraint.String:
			input[key] = []byte("'" + val.String() + "'")
		default:
			input[key] = []byte(val.String())
		}
	}

	ret := t.listener.GetComputeResults(input)

	out := make(map[string]constraint.Node)
	for key, val := range ret {
		out[key] = constraint.NewNode(val)
	}

	return out, nil
}
