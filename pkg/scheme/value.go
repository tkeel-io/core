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

package scheme

import (
	"encoding/json"

	logf "github.com/tkeel-io/core/pkg/logfield"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/util"
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
	Name              string                 `json:"name" mapstructure:"name"`
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
			return index, cfg, xerrors.ErrPatchTypeInvalid
		}

		define := cfg.getStructDefine()
		c, ok := define.Fields[segs[index]]
		if !ok {
			return index, cfg, xerrors.ErrPatchPathLack
		}

		cc := &c
		return cc.getConfig(segs, index+1)
	}
	return index, cfg, nil
}

func (cfg *Config) AppendField(c Config) error {
	if cfg.Type != PropertyTypeStruct {
		return xerrors.ErrInvalidNodeType
	}
	define := cfg.getStructDefine()
	define.Fields[c.ID] = c
	return nil
}

func (cfg *Config) RemoveField(id string) error {
	if cfg.Type != PropertyTypeStruct {
		return xerrors.ErrInvalidNodeType
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

func Parse(bytes []byte) (map[string]*Config, error) {
	// parse state config again.
	configs := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &configs); nil != err {
		log.L().Error("json unmarshal", logf.Error(err), logf.String("configs", string(bytes)))
		return nil, errors.Wrap(err, "json unmarshal")
	}

	var err error
	var cfg Config
	cfgs := make(map[string]*Config)
	for key, val := range configs {
		if cfg, err = ParseConfigFrom(val); nil != err {
			// TODO: dispose error.
			log.L().Error("parse configs", logf.Error(err))
			continue
		}
		cfgs[key] = &cfg
	}

	return cfgs, nil
}

func ParseFrom(bytes []byte) (*Config, error) {
	v := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &v); nil != err {
		log.L().Error("unmarshal Config", logf.Error(err))
		return nil, errors.Wrap(err, "unmarshal Config")
	}

	cfg, err := ParseConfigFrom(v)
	return &cfg, errors.Wrap(err, "parse Config")
}

func ParseConfigFrom(data interface{}) (cfg Config, err error) {
	cfgRequest := Config{}
	if err = mapstructure.Decode(data, &cfgRequest); nil != err {
		return cfg, errors.Wrap(err, "decode property config failed")
	} else if cfgRequest, err = parseField(cfgRequest); nil != err {
		return cfg, errors.Wrap(err, "parse config  failed")
	}
	return cfgRequest, nil
}

func parseField(in Config) (out Config, err error) {
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
			return out, xerrors.ErrEntityConfigInvalid
		}
		arrDefine.ElemType, err = parseField(arrDefine.ElemType)
		in.Define["elem_type"] = arrDefine.ElemType
	case PropertyTypeStruct:
		jsonDefine, jsonDefine2 := newDefineStruct(), newDefineStruct()
		if err = mapstructure.Decode(in.Define, &jsonDefine); nil != err {
			return out, errors.Wrap(err, "parse property config failed")
		}

		for cfgID, field := range jsonDefine.Fields {
			var cfg Config
			if cfg, err = parseField(field); nil != err {
				return out, errors.Wrap(err, "parse property config failed")
			}
			cfg.ID = cfgID
			jsonDefine2.Fields[cfgID] = cfg
		}

		in.Define["fields"] = jsonDefine2.Fields
	default:
		return out, xerrors.ErrEntityConfigInvalid
	}

	in.LastTime = lastTimestamp(in.LastTime)
	return in, errors.Wrap(err, "parse property config failed")
}

func lastTimestamp(timestamp int64) int64 {
	if timestamp == 0 {
		timestamp = util.UnixMilli()
	}
	return timestamp
}
