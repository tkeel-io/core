package constraint

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	PropertyTypeInt    = "int"
	PropertyTypeBool   = "bool"
	PropertyTypeFloat  = "float"
	PropertyTypeDouble = "double"
	PropertyTypeString = "string"
	PropertyTypeArray  = "array"
	PropertyTypeStruct = "struct"

	DefineFieldArrayLength  = "length"
	DefineFieldArrayElemCfg = "elem_type"
	DefineFieldStructFields = "fields"
)

type Config struct {
	item

	ID                string                 `json:"id"`
	Type              string                 `json:"type"`      // 用于描述entity运行时的属性值的结构信息.
	DataType          string                 `json:"data_type"` // 用于描述entity运行时属性值的存在形式，默认[]byte.
	Weight            int                    `json:"weight"`
	Enabled           bool                   `json:"enabled"`
	EnabledSearch     bool                   `json:"enabled_search"`
	EnabledTimeSeries bool                   `json:"enabled_time_series"`
	Description       string                 `json:"description"`
	Define            map[string]interface{} `json:"define"`
	LastTime          int64                  `json:"last_time"`
}

func (cfg Config) getArrayDefine() DefineArray {
	var arrDefine DefineArray
	if length, ok := cfg.Define[DefineFieldArrayLength].(int); !ok {
		arrDefine.Length = length
	}
	if elemT, ok := cfg.Define[DefineFieldArrayElemCfg].(Config); !ok {
		arrDefine.ElemType = elemT
	}
	return arrDefine
}

func (cfg Config) getStructDefine() DefineStruct {
	var jsonDefine DefineStruct
	if fields, ok := cfg.Define[DefineFieldStructFields].([]Config); !ok {
		jsonDefine.Fields = fields
	}
	return jsonDefine
}

type DefineStruct struct {
	Fields []Config `json:"fields"`
}

type DefineArray struct {
	Length   int    `json:"length"`
	ElemType Config `json:"elem_type"`
}

func ParseConfigsFrom(data map[string]interface{}) (cfg Config, err error) {
	cfgRequest := Config{}
	if err = mapstructure.Decode(data, &cfgRequest); nil != err {
		return cfg, errors.Wrap(err, "parse property config failed")
	} else if cfgRequest, err = parseField(cfgRequest); nil != err {
		return cfg, errors.Wrap(err, "parse property config failed")
	}
	return cfgRequest, nil
}

func parseField(in Config) (out Config, err error) { //nolint
	switch in.Type {
	case PropertyTypeInt:
	case PropertyTypeBool:
	case PropertyTypeFloat:
	case PropertyTypeDouble:
	case PropertyTypeString:
	case PropertyTypeArray:
		arrDefine := DefineArray{}
		if err = mapstructure.Decode(in.Define, &arrDefine); nil != err {
			return out, errors.Wrap(err, "parse property config failed")
		} else if arrDefine.Length <= 0 {
			return out, ErrEntityConfigInvalid
		}

		arrDefine.ElemType, err = parseField(arrDefine.ElemType)
		in.Define["elem_type"] = arrDefine.ElemType
	case PropertyTypeStruct:
		jsonDefine, jsonDefine2 := DefineStruct{}, DefineStruct{}
		if err = mapstructure.Decode(in.Define, &jsonDefine); nil != err {
			return out, errors.Wrap(err, "parse property config failed")
		}

		for _, field := range jsonDefine.Fields {
			var cfg Config
			if cfg, err = parseField(field); nil != err {
				return out, errors.Wrap(err, "parse property config failed")
			}
			jsonDefine2.Fields = append(jsonDefine2.Fields, cfg)
		}

		in.Define["fields"] = jsonDefine2.Fields
	default:
		return out, ErrEntityConfigInvalid
	}
	return in, errors.Wrap(err, "parse property config failed")
}
