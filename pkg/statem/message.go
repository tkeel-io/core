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

package statem

import (
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
)

// PropertyMessage state property message.
type StateMessage struct {
	MessageBase

	StateID  string `json:"state_id"`
	Operator string `json:"operator"`
}

func NewPropertyMessage(id string, props map[string]constraint.Node) PropertyMessage {
	return PropertyMessage{
		StateID:    id,
		Operator:   "replace",
		Properties: props,
	}
}

// PropertyMessage state property message.
type PropertyMessage struct {
	MessageBase

	StateID    string                     `json:"state_id"`
	Operator   string                     `json:"operator"`
	Properties map[string]constraint.Node `json:"properties"`
}

type MapperMessage struct {
	MessageBase

	Operator string     `json:"operator"`
	Mapper   MapperDesc `json:"mapper"`
}

type TentacleMsg struct {
	MessageBase

	Operator string            `json:"operator"`
	StateID  string            `json:"state_id"`
	Items    []mapper.WatchKey `json:"items"`
}
