package dao

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"

	msgpack "github.com/shamaton/msgpack/v2"
)

type entityCodec struct {
}

func (ec entityCodec) Key(id string) string {
	return EntityStorePrefix + id
}

func (ec entityCodec) Encode(en *Entity) ([]byte, error) {
	bytes, err := msgpack.Marshal(en)
	return bytes, errors.Wrap(err, "encode entity")
}

func (ec entityCodec) Decode(data []byte, en *Entity) error {
	var v = make(map[string]interface{})
	if err := msgpack.Unmarshal(data, &v); nil != err {
		return errors.Wrap(err, "decode entity")
	}

	// decode entity.
	if err := mapstructure.Decode(v, &en); nil != err {
		return errors.Wrap(err, "mapstructure entity")
	}

	switch properties := v["properties"].(type) {
	case nil:
	case map[string]interface{}:
		en.Properties = make(map[string]constraint.Node)
		for key, val := range properties {
			en.Properties[key] = constraint.NewNode(val)
		}
	case map[interface{}]interface{}:
		en.Properties = make(map[string]constraint.Node)
		for key, val := range properties {
			keyString, _ := key.(string)
			en.Properties[keyString] = constraint.NewNode(val)
		}
	default:
		return errors.Wrap(xerrors.ErrInvalidProperties, "should be map[interface{}]interface{} or map[string]interface{}")
	}

	return nil
}
