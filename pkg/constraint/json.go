package constraint

import (
	"fmt"
	"reflect"

	"strconv"
	"strings"
)

var (
	UNDEFINED_RESULT = &DefaultNode{typ: Undefined}
	NULL_RESULT      = &DefaultNode{typ: Undefined}
)

// Type node type
type Type int

const (
	// Undefine is Not a value
	// This isn't explicitly representable in JSON except by omitting the value.
	Undefined Type = iota
	// Null is a null json value
	Null
	// Bool is a json boolean
	Bool
	// Number is json number, include Int and Float
	Number
	// Int is json number, a discrete Int
	Integer
	// Float is json number
	Float
	// String is a json string
	String
	// JSON is a raw block of JSON
	JSON
	// RAW for golang runtine.
	RAW
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

//Node interface
type Node interface {
	Type() Type
	To(Type) Node
	String() string
}

//DefaultNode interface
type DefaultNode struct {
	// Type is the json type
	typ Type
	// raw is the raw json
	raw string
}

func (r DefaultNode) Type() Type { return r.typ }
func (r DefaultNode) To(Type) Node {
	return r
}
func (r DefaultNode) String() string {
	return r.raw
}

type BoolNode bool

func (r BoolNode) Type() Type { return Bool }
func (r BoolNode) To(typ Type) Node {
	switch typ {
	case Bool:
		return r
	case String:
		return StringNode(fmt.Sprintf("%t", r))
	}
	return UNDEFINED_RESULT
}
func (r BoolNode) String() string {
	return fmt.Sprintf("%t", r)
}

type IntNode int64

func (r IntNode) Type() Type { return Integer }
func (r IntNode) To(typ Type) Node {
	switch typ {
	case Number, Integer:
		return r
	case Float:
		return FloatNode(r)
	case String:
		return StringNode(fmt.Sprintf("%d", r))
	}
	return UNDEFINED_RESULT
}
func (r IntNode) String() string {
	return fmt.Sprintf("%d", r)
}

type FloatNode float64

func (r FloatNode) Type() Type { return Float }
func (r FloatNode) To(typ Type) Node {
	switch typ {
	case Number, Float:
		return r
	case Integer:
		return IntNode(r)
	case String:
		return StringNode(strconv.FormatFloat(float64(r), 'f', -1, 64))
	}
	return UNDEFINED_RESULT
}
func (r FloatNode) String() string {
	return fmt.Sprintf("%f", r)
}

type StringNode string

func (r StringNode) Type() Type { return String }
func (r StringNode) To(typ Type) Node {
	switch typ {
	case String:
		return r
	case Bool:
		b, err := strconv.ParseBool(string(r))
		if err != nil {
			return UNDEFINED_RESULT
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
			return UNDEFINED_RESULT
		}
		return IntNode(b)
	case Float:
		b, err := strconv.ParseFloat(string(r), 64)
		if err != nil {
			return UNDEFINED_RESULT
		}
		return FloatNode(b)
	}
	return UNDEFINED_RESULT
}
func (r StringNode) String() string {
	return string(r)
}

// JSONNode maybe Object or Array
type JSONNode string

func (r JSONNode) Type() Type { return JSON }
func (r JSONNode) To(typ Type) Node {
	return UNDEFINED_RESULT
}

func (r JSONNode) String() string {
	return string(r)
}

type RawNode []byte

func (r RawNode) Type() Type { return RAW }
func (r RawNode) To(tp Type) Node {
	switch tp {
	case String:
		return StringNode(r)
	case JSON:
		return JSONNode(r)
	default:
		return UNDEFINED_RESULT
	}
}

func (r RawNode) String() string {
	return string(r)
}

func NewNode(v interface{}) Node {
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
		return RawNode(val)
	case bool:
		return BoolNode(val)
	default:
		if reflect.Ptr == reflect.TypeOf(val).Kind() {
			// deference pointer.
			return NewNode(reflect.ValueOf(val).Elem().Interface())
		}
		return RawNode(fmt.Sprintf("%v", val))
	}
}
