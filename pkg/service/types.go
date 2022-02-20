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

package service

import (
	"encoding/json"

	apim "github.com/tkeel-io/core/pkg/manager"
)

type Entity = apim.Base

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderOwner       = "Owner"
	HeaderType        = "Type"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	User   = "user_id"
)

type Marshalable interface {
	String() string
}

type MarshalField struct {
	val interface{}
}

func (m MarshalField) String() string {
	bytes, _ := json.Marshal(m.val)
	return string(bytes)
}

func NewMarField(v interface{}) MarshalField {
	return MarshalField{val: v}
}

type PatchData struct {
	Path     string
	Operator string
	Value    interface{}
}
