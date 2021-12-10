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

package service

import (
	"encoding/json"
	"errors"

	"github.com/tkeel-io/core/pkg/statem"

	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api.service")

var (
	ErrEntityMapperNil       = errors.New("mapper is nil")
	ErrEntityConfigInvalid   = errors.New("entity config format invalid")
	ErrEntityInvalidParams   = errors.New("invalid params")
	ErrEntityEmptyRequest    = errors.New("empty request")
	ErrEntityPropertyIDEmpty = errors.New("emtpty property id")
)

type Entity = statem.Base

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderOwner       = "Owner"
	HeaderType        = "Type"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	User   = "user_id"
)

/*
	ErrorIf 用起来是挺爽的，但是会存在一个问题，那就是我们的日志除了人眼分析， 更多的时候是需要为日志分析系统提供数据源的，而日志分析系统对日志数据的约束是json-object.
*/

func ErrorIf(err *error, fmtString string, args ...interface{}) {
	if nil != *err {
		for index, arg := range args {
			if val, ok := arg.(Marshalable); ok {
				args[index] = val.String()
			}
		}
		log.Errorf(fmtString, *err, args)
	}
}

type Marshalable interface {
	String() string
}

type MarshalField struct {
	val interface{}
}

func (m MarshalField) String() string {
	bytes, _ := json.Marshal(m.val)
	return string(bytes)
}

func NewMarField(v interface{}) MarshalField {
	return MarshalField{val: v}
}
