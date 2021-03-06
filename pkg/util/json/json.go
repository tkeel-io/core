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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	"github.com/tkeel-io/tdtl"
)

func EncodeJSON(kvalues map[string]tdtl.Node) ([]byte, error) {
	collect := collectjs.New("{}")
	for key, val := range kvalues {
		collect.Set(key, val.Raw())
	}
	return collect.GetRaw(), errors.Wrap(collect.GetError(), "Encode Json")
}

func EncodeJSONZ(kvalues map[string]tdtl.Node) ([]byte, error) {
	msgArr := []string{}
	for key, val := range kvalues {
		msgArr = append(msgArr,
			fmt.Sprintf("\"%s\":%s", key, string(val.Raw())))
	}

	return []byte(fmt.Sprintf("{%s}", strings.Join(msgArr, ","))), nil
}

func NewNode(dataType jsonparser.ValueType, value []byte) tdtl.Node {
	switch dataType {
	case jsonparser.String:
		return tdtl.New(value)
	case jsonparser.Number:
		return tdtl.New(value).To(tdtl.Number)
	case jsonparser.Object:
		return tdtl.New(value)
	case jsonparser.Array:
		return tdtl.New(value)
	case jsonparser.Boolean:
		return tdtl.New(value).To(tdtl.Bool)
	case jsonparser.Null:
		return tdtl.NULL_RESULT
	default:
		return tdtl.UNDEFINED_RESULT
	}
}
