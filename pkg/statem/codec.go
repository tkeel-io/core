package statem

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"

	msgpack "github.com/shamaton/msgpack/v2"
)

func EncodeBase(base *Base) ([]byte, error) {
	bytes, err := json.Marshal(base.Configs)
	if nil != err {
		log.Error("encode Base", zap.Any("base", base), zap.Error(err))
		return nil, errors.Wrap(err, "encode Base")
	}

	log.Debug("encode Base", logger.EntityID(base.ID), zap.String("configs", string(bytes)))

	// set base.ConfigsBytes
	base.ConfigsBytes = bytes
	bytes, err = msgpack.Marshal(base)
	return bytes, errors.Wrap(err, "encode Base")
}

func DecodeBase(data []byte) (*Base, error) { //nolint
	var v = make(map[string]interface{})
	if err := msgpack.Unmarshal(data, &v); nil != err {
		return nil, errors.Wrap(err, "decode Base-State json")
	}

	// unmarshal Configs.
	var configs = make(map[string]constraint.Config)
	configsBytes, _ := v["configs_bytes"].([]byte)
	if err := json.Unmarshal(configsBytes, &configs); nil != err {
		log.Error("decode Base", zap.Error(err), zap.Any("data", string(data)))
		// return nil, errors.Wrap(err, "decode Base")
	}

	var base Base
	base.Configs = make(map[string]constraint.Config)
	for key, value := range configs {
		cfg, err := constraint.ParseConfigsFrom(value)
		if nil != err {
			continue
		}
		base.Configs[key] = cfg
	}

	// decode base.
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
			keyString, _ := key.(string)
			base.KValues[keyString] = constraint.NewNode(val)
		}
	default:
		return nil, ErrInvalidProperties
	}

	return &base, nil
}
