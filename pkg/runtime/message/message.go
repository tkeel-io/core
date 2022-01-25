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

package message

import (
	"github.com/tkeel-io/core/pkg/constraint"
)

type Method string

const (
	// enumerate method.
	SMMethodSetConfigs   Method = "SetConfigs"
	SMMethodPatchConfigs Method = "PatchConfigs"
	SMMethodDeleteEntity Method = "DeleteEntity"

	// enumerate PropertyMessage fields.
	FiledStateID    = "state_id"
	FieldOperator   = "operator"
	FieldProperties = "properties"
)

func NewPropertyMessage(id string, props map[string]constraint.Node) PropertyMessage {
	return PropertyMessage{
		StateID:    id,
		Operator:   "replace",
		Properties: props,
	}
}

// 用于操作实体属性的消息.
type PropertyMessage struct {
	MessageBase

	StateID    string                     `json:"state_id"`
	Operator   string                     `json:"operator"`
	Properties map[string]constraint.Node `json:"properties"`
}

// 用于操作实体属性的消息, 消息处理后自动Flush.
type FlushPropertyMessage PropertyMessage

// 用于操作实体配置的消息.
//		1. 属性配置.
//		2. 实体管理.
type StateMessage struct {
	MessageBase

	StateID string      `json:"state_id"`
	Method  Method      `json:"method"`
	Value   interface{} `json:"value"`
}

// Flusher interface for State Machine Message.
type Flusher interface{ viod() }

func (s StateMessage) viod()         {}
func (f FlushPropertyMessage) viod() {}

func (s StateMessage) String() string         { return "StateMessage" }
func (s PropertyMessage) String() string      { return "PropertyMessage" }
func (f FlushPropertyMessage) String() string { return "FlushPropertyMessage" }
