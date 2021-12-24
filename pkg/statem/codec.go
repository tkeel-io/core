package statem

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"

	msgpack "github.com/shamaton/msgpack/v2"
)

func EncodeBase(base *Base) ([]byte, error) {
	bytes, err := msgpack.Marshal(base)
	return bytes, errors.Wrap(err, "encode Base")
}

func DecodeBase(data []byte) (*Base, error) {
	var v = make(map[string]interface{})
	if err := msgpack.Unmarshal(data, &v); nil != err {
		return nil, errors.Wrap(err, "decode Base-State json")
	}

	var base Base
	if err := mapstructure.Decode(v, &base); nil != err {
		return nil, errors.Wrap(err, "decode Base-State struct")
	}

	switch properties := v["properties"].(type) {
	case nil:
	case map[string]interface{}:
		base.KValues = make(map[string]constraint.Node)
		for key, val := range properties {
			base.KValues[key] = constraint.NewNode(val)
		}
	case map[interface{}]interface{}:
		base.KValues = make(map[string]constraint.Node)
		for key, val := range properties {
			base.KValues[key.(string)] = constraint.NewNode(val)
		}
	default:
		return nil, ErrInvalidProperties
	}
	return &base, nil
}
