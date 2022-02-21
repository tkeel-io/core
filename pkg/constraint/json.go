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
	"encoding/json"
	"fmt"
	"reflect"

	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

var (
	UndefineResult = &DefaultNode{typ: Undefined}
	NullResult     = &DefaultNode{typ: Undefined}
)

// Type node type.
type Type int

const (
	// Undefine is Not a value.
	// This isn't explicitly representable in JSON except by omitting the value.
	Undefined Type = iota
	// Null is a null json value.
	Null
	// Bool is a json boolean.
	Bool
	// Number is json number, include Int and Float.
	Number
	// Int is json number, a discrete Int.
	Integer
	// Float is json number.
	Float
	// String is a json string.
	String
	// Array is a json array.
	Array
	// JSON is a raw block of JSON.
	JSON
)

// String returns a string representation of the type.
func (t Type) String() string {
	switch t {
	default:
		return "Undefined"
	case Null:
		return "Null"
	case Bool:
		return "Bool"
	case Integer:
		return "Integer"
	case Float:
		return "Float"
	case String:
		return "String"
	case Array:
		return "Array"
	case JSON:
		return "JSON"
	}
}

// Node interface.
type Node interface {
	Type() Type
	To(Type) Node
	Copy() Node
	String() string
	Value() interface{}
}

// DefaultNode interface.
type DefaultNode struct {
	// Type is the json type.
	typ Type
	// raw is the raw json.
	raw string
}

func (r DefaultNode) Type() Type         { return r.typ }
func (r DefaultNode) To(Type) Node       { return r }
func (r DefaultNode) Copy() Node         { return DefaultNode{typ: r.typ, raw: r.raw} }
func (r DefaultNode) String() string     { return r.raw }
func (r DefaultNode) Value() interface{} { return r.raw }

type BoolNode bool

func (r BoolNode) String() string     { return fmt.Sprintf("%t", r) }
func (r BoolNode) Value() interface{} { return bool(r) }
func (r BoolNode) Type() Type         { return Bool }
func (r BoolNode) Copy() Node         { return r }
func (r BoolNode) To(typ Type) Node {
	switch typ {
	case Bool:
		return r
	case String:
		return StringNode(fmt.Sprintf("%t", r))
	default:
		return UndefineResult
	}
}

type IntNode int64

func (r IntNode) Type() Type         { return Integer }
func (r IntNode) String() string     { return strconv.FormatInt(int64(r), 10) }
func (r IntNode) Value() interface{} { return int64(r) }
func (r IntNode) Copy() Node         { return r }
func (r IntNode) To(typ Type) Node {
	switch typ {
	case Number, Integer:
		return r
	case Float:
		return FloatNode(r)
	case String:
		return StringNode(r.String())
	default:
		return UndefineResult
	}
}

type FloatNode float64

func (r FloatNode) Type() Type         { return Float }
func (r FloatNode) String() string     { return strconv.FormatFloat(float64(r), 'f', -1, 64) }
func (r FloatNode) Value() interface{} { return float64(r) }
func (r FloatNode) Copy() Node         { return r }
func (r FloatNode) To(typ Type) Node {
	switch typ {
	case Number, Float:
		return r
	case Integer:
		return IntNode(r)
	case String:
		return StringNode(strconv.FormatFloat(float64(r), 'f', -1, 64))
	default:
		return UndefineResult
	}
}

type StringNode string

func (r StringNode) Type() Type         { return String }
func (r StringNode) String() string     { return string(r) }
func (r StringNode) Value() interface{} { return string(r[1 : len(r)-1]) }
func (r StringNode) Copy() Node {
	res := make([]byte, len(r))
	copy(res, r)
	return StringNode(res)
}

func (r StringNode) To(typ Type) Node {
	switch typ {
	case String:
		return r
	case Bool:
		b, err := strconv.ParseBool(string(r))
		if err != nil {
			return UndefineResult
		}
		return BoolNode(b)
	case Number:
		if !strings.Contains(string(r), ".") {
			return r.To(Integer)
		}
		return r.To(Float)
	case Integer:
		b, err := strconv.ParseInt(string(r), 10, 64)
		if err != nil {
			return UndefineResult
		}
		return IntNode(b)
	case Float:
		b, err := strconv.ParseFloat(string(r), 64)
		if err != nil {
			return UndefineResult
		}
		return FloatNode(b)
	default:
		return UndefineResult
	}
}

func Unwrap(s Node) string {
	switch s.Type() {
	case String:
		ss, _ := s.Value().(string)
		return ss
	default:
		return s.String()
	}
}

type NullNode struct{}

func (r NullNode) Type() Type         { return Null }
func (r NullNode) String() string     { return "null" }
func (r NullNode) Value() interface{} { return nil }
func (r NullNode) Copy() Node         { return r }
func (r NullNode) To(typ Type) Node {
	switch typ {
	case Null:
		return r
	case JSON:
		return JSONNode("{}")
	case Array:
		return ArrayNode("[]")
	default:
		return UndefineResult
	}
}

type ArrayNode []byte

func (r ArrayNode) Type() Type     { return Array }
func (r ArrayNode) String() string { return string(r) }
func (r ArrayNode) Value() interface{} {
	var data interface{}
	_ = json.Unmarshal(r, &data)
	return data
}
func (r ArrayNode) Copy() Node {
	res := make([]byte, len(r))
	copy(res, r)
	return ArrayNode(res)
}

func (r ArrayNode) To(typ Type) Node {
	switch typ {
	case String:
		return StringNode(r)
	case Array:
		return r
	case JSON:
		return JSONNode(r)
	default:
		return UndefineResult
	}
}

// JSONNode maybe Object or Array.
type JSONNode []byte

func (r JSONNode) Type() Type     { return JSON }
func (r JSONNode) String() string { return "\"" + string(r) + "\"" }
func (r JSONNode) Value() interface{} {
	var data interface{}
	_ = json.Unmarshal(r, &data)
	return data
}

func (r JSONNode) Copy() Node {
	res := make([]byte, len(r))
	copy(res, r)
	return JSONNode(res)
}

func (r JSONNode) To(typ Type) Node {
	switch typ {
	case String:
		return StringNode(r)
	case JSON:
		return r
	default:
		return UndefineResult
	}
}

func NewNode(v interface{}) Node {
	switch val := v.(type) {
	case float32:
		return FloatNode(val)
	case float64:
		return FloatNode(val)
	case uint8, int8, uint16, int16, uint,
		int, uint32, int32, int64, uint64:
		return StringNode(fmt.Sprintf("%v", val)).To(Integer)
	case string:
		return StringNode(val)
	case []byte:
		return JSONNode(val)
	case bool:
		return BoolNode(val)
	case map[string]interface{}:
		data, _ := json.Marshal(v)
		return JSONNode(string(data))
	case nil:
		return NullNode{}
	default:
		valKind := reflect.TypeOf(val).Kind()
		if reflect.Ptr == valKind {
			// deference pointer.
			return NewNode(reflect.ValueOf(val).Elem().Interface())
		} else if reflect.Slice == valKind {
			data, _ := json.Marshal(v)
			return JSONNode(string(data))
		}

		return UndefineResult
	}
}

func ToBytesWithWrapString(val Node) []byte {
	if nil == val {
		return []byte{}
	}

	switch val.Type() {
	case JSON:
		jsonVal, _ := val.(JSONNode)
		return []byte(jsonVal)
	case String:
		return []byte(val.String())
	default:
		return []byte(val.String())
	}
}

func EncodeJSON(kvalues map[string]Node) ([]byte, error) {
	collect := collectjs.New("{}")
	for key, val := range kvalues {
		collect.Set(key, []byte(val.String()))
	}
	return collect.GetRaw(), errors.Wrap(collect.GetError(), "Encode Json")
}

func DecodeJSON(values []byte) (map[string]Node, error) {
	var result = make(map[string]Node)
	collect := collectjs.ByteNew(values)
	collect.Foreach(func(key, value []byte) {
		keyString := string(key)
		ct := collectjs.ByteNew(value)
		switch ct.GetDataType() {
		case jsonparser.String.String():
			result[keyString] = StringNode("\"" + string(value) + "\"")
		case jsonparser.Number.String():
			result[keyString] = StringNode(value)
		case jsonparser.Object.String():
			result[keyString] = JSONNode(value)
		case jsonparser.Array.String():
			result[keyString] = ArrayNode(value)
		case jsonparser.Boolean.String():
			flag, _ := jsonparser.ParseBoolean(value)
			result[keyString] = NewNode(flag)
		case jsonparser.Null.String():
			result[keyString] = NullNode{}
		default:
			log.Error("invalid json type",
				zap.Error(jsonparser.UnknownValueTypeError))
		}
	})
	return result, nil
}
