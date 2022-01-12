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
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
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
	ID                string                 `json:"id" mapstructure:"id"`
	Type              string                 `json:"type" mapstructure:"type"`
	Weight            int                    `json:"weight" mapstructure:"weight"`
	Enabled           bool                   `json:"enabled" mapstructure:"enabled"`
	EnabledSearch     bool                   `json:"enabled_search" mapstructure:"enabled_search"`
	EnabledTimeSeries bool                   `json:"enabled_time_series" mapstructure:"enabled_time_series"`
	Description       string                 `json:"description" mapstructure:"description"`
	Define            map[string]interface{} `json:"define" mapstructure:"define"`
	LastTime          int64                  `json:"last_time" mapstructure:"last_time"`
}

func (cfg *Config) getArrayDefine() DefineArray {
	length, _ := cfg.Define[DefineFieldArrayLength].(int)
	etype, _ := cfg.Define[DefineFieldArrayElemCfg].(Config)
	return DefineArray{Length: length, ElemType: etype}
}

func (cfg *Config) getStructDefine() DefineStruct {
	fields, ok := cfg.Define[DefineFieldStructFields].(map[string]Config)
	if !ok {
		fields = make(map[string]Config)
		cfg.Define[DefineFieldStructFields] = fields
	}
	return DefineStruct{Fields: fields}
}

func (cfg *Config) GetConfig(segs []string, index int) (int, *Config, error) {
	return cfg.getConfig(segs, index)
}

func (cfg *Config) getConfig(segs []string, index int) (int, *Config, error) {
	if len(segs) > index {
		if cfg.Type != PropertyTypeStruct {
			return index, cfg, ErrPatchTypeInvalid
		}

		define := cfg.getStructDefine()
		c, ok := define.Fields[segs[index]]
		if !ok {
			return index, cfg, ErrPatchPathLack
		}

		cc := &c
		return cc.getConfig(segs, index+1)
	}
	return index, cfg, nil
}

func (cfg *Config) AppendField(c Config) error {
	if cfg.Type != PropertyTypeStruct {
		return ErrInvalidNodeType
	}
	define := cfg.getStructDefine()
	define.Fields[c.ID] = c
	return nil
}

func (cfg *Config) RemoveField(id string) error {
	if cfg.Type != PropertyTypeStruct {
		return ErrInvalidNodeType
	}
	define := cfg.getStructDefine()
	delete(define.Fields, id)
	return nil
}

type DefineStruct struct {
	Fields map[string]Config `json:"fields" mapstructure:"fields"`
}

func newDefineStruct() DefineStruct {
	return DefineStruct{Fields: make(map[string]Config)}
}

type DefineArray struct {
	Length   int    `json:"length" mapstructure:"length"`
	ElemType Config `json:"elem_type" mapstructure:"elem_type"`
}

func ParseConfigsFrom(data interface{}) (cfg Config, err error) {
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
		jsonDefine, jsonDefine2 := newDefineStruct(), newDefineStruct()
		if err = mapstructure.Decode(in.Define, &jsonDefine); nil != err {
			return out, errors.Wrap(err, "parse property config failed")
		}

		for _, field := range jsonDefine.Fields {
			var cfg Config
			if cfg, err = parseField(field); nil != err {
				return out, errors.Wrap(err, "parse property config failed")
			}
			jsonDefine2.Fields[cfg.ID] = cfg
		}

		in.Define["fields"] = jsonDefine2.Fields
	default:
		log.Info("===================", in)
		return out, ErrEntityConfigInvalid
	}

	return in, errors.Wrap(err, "parse property config failed")
}
