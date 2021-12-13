package statem

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
)

func EncodeBase(base *Base) ([]byte, error) {
	return json.Marshal(base)
}

func DecodeBase(data []byte) (*Base, error) {
	var v = make(map[string]interface{})
	if err := json.Unmarshal(data, &v); nil != err {
		return nil, errors.Wrap(err, "decode Base-State json")
	}

	var base Base
	if err := mapstructure.Decode(v, &base); nil != err {
		return nil, errors.Wrap(err, "decode Base-State struct")
	}

	if properties, ok := v["properties"].(map[string]interface{}); ok {
		base.KValues = make(map[string]constraint.Node)
		for key, val := range properties {
			base.KValues[key] = constraint.NewNode(val)
		}
	}

	return &base, nil
}
