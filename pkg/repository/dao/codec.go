package dao

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"

	msgpack "github.com/shamaton/msgpack/v2"
)

func GetEntityCodec() entityCodec { //nolint
	return entityCodec{}
}

type entityCodec struct {
}

func (ec entityCodec) Key(id string) string {
	return EntityStorePrefix + id
}

func (ec entityCodec) Encode(en *Entity) (bytes []byte, err error) {
	bytes = []byte(`{}`)
	for key, val := range en.Properties {
		if bytes, err = collectjs.Set(bytes, key, []byte(val.String())); nil != err {
			return bytes, errors.Wrap(err, "patch replace")
		}
	}

	en.PropertyBytes = bytes
	bytes, err = msgpack.Marshal(en)

	// reset.
	en.PropertyBytes = nil
	return bytes, errors.Wrap(err, "encode entity")
}

func (ec entityCodec) Decode(data []byte, en *Entity) error {
	if err := msgpack.Unmarshal(data, en); nil != err {
		return errors.Wrap(err, "decode entity")
	}

	en.Properties = make(map[string]tdtl.Node)
	collectjs.ForEach(en.PropertyBytes, jsonparser.Object,
		func(key, value []byte, dataType jsonparser.ValueType) {
			en.Properties[string(key)] = xjson.NewNode(dataType, value)
		})

	// reset .
	en.PropertyBytes = nil
	return nil
}
