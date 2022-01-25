package message

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/constraint"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

type PropsMessageCodec interface {
	Encode(PropertyMessage) ([]byte, error)
	Decode([]byte) (PropertyMessage, error)
}

type propsMessageCodec struct{}

var defaultPropsCodec = propsMessageCodec{}

func GetPropsCodec() PropsMessageCodec {
	return &defaultPropsCodec
}

func (c *propsMessageCodec) Encode(msg PropertyMessage) ([]byte, error) {
	var err error
	bytes := []byte("{}")

	// encode properties.
	propBytes, err := constraint.EncodeJSON(msg.Properties)
	if nil != err {
		return bytes, errors.Wrap(err, "encode entity properties")
	}

	// construct message fields.
	if bytes, err = collectjs.Set(bytes, FiledStateID, wrapString(msg.StateID)); nil != err {
		return bytes, errors.Wrap(err, "encode entity state_id")
	} else if bytes, err = collectjs.Set(bytes, FieldOperator, wrapString(msg.Operator)); nil != err {
		return bytes, errors.Wrap(err, "encode entity operator")
	} else if bytes, err = collectjs.Set(bytes, FieldProperties, propBytes); nil != err {
		return bytes, errors.Wrap(err, "encode entity operator")
	}

	return bytes, nil
}

func (c *propsMessageCodec) Decode(bytes []byte) (msg PropertyMessage, err error) {
	msg.Properties = make(map[string]constraint.Node)
	cc := collectjs.ByteNew(bytes)
	cc.Foreach(func(key []byte, value []byte) {
		switch string(key) {
		case FiledStateID:
			msg.StateID = unwrapString(value)
		case FieldOperator:
			msg.Operator = unwrapString(value)
		case FieldProperties:
			ccc := collectjs.ByteNew(value)
			ccc.Foreach(func(key []byte, value []byte) {
				msg.Properties[string(key)] = constraint.NewNode(value)
			})
		default:
			msg.Properties[string(key)] = constraint.NewNode(value)
		}
	})

	if msg.Operator == "" {
		// default operator for pubsub.
		msg.Operator = constraint.PatchOpReplace.String()
	}

	return msg, errors.Wrap(err, "decode property message")
}

func wrapString(s string) []byte {
	return []byte("\"" + s + "\"")
}

func unwrapString(bytes []byte) string {
	if len(bytes) > 2 {
		return string(bytes[1 : len(bytes)-1])
	}
	log.Warn("unwrap string failed", zfield.Value(string(bytes)))
	return string(bytes)
}
