package constraint

import (
	"encoding/json"
	"fmt"
	"reflect"

	"strconv"
	"strings"
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
	case JSON:
		return "JSON"
	}
}

// Node interface.
type Node interface {
	Type() Type
	To(Type) Node
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
func (r DefaultNode) String() string     { return r.raw }
func (r DefaultNode) Value() interface{} { return r.raw }

type BoolNode bool

func (r BoolNode) String() string     { return fmt.Sprintf("%t", r) }
func (r BoolNode) Value() interface{} { return bool(r) }
func (r BoolNode) Type() Type         { return Bool }
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
func (r StringNode) Value() interface{} { return string(r) }
func (r StringNode) To(typ Type) Node { //nolint
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

type NullNode struct{}

func (r NullNode) Type() Type         { return Null }
func (r NullNode) String() string     { return "null" }
func (r NullNode) Value() interface{} { return nil }
func (r NullNode) To(typ Type) Node {
	switch typ {
	case Null:
		return r
	default:
		return UndefineResult
	}
}

// JSONNode maybe Object or Array.
type JSONNode []byte

func (r JSONNode) Type() Type     { return JSON }
func (r JSONNode) String() string { return string(r) }
func (r JSONNode) Value() interface{} {
	var data interface{}
	_ = json.Unmarshal(r, &data)
	return data
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

func NewNode(v interface{}) Node { //nolint
	switch val := v.(type) {
	case float32:
		return FloatNode(val)
	case float64:
		return FloatNode(val)
	case uint8, int8, uint16, int16, uint, int, uint32, int32, int64, uint64:
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
		if reflect.Ptr == reflect.TypeOf(val).Kind() {
			// deference pointer.
			return NewNode(reflect.ValueOf(val).Elem().Interface())
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
		return []byte("\"" + val.String() + "\"")
	default:
		return []byte(val.String())
	}
}
