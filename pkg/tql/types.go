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

package tql

import "github.com/tkeel-io/core/pkg/constraint"

/*
   1. 对MQL的静态分析
   	1. 解析Target Entity
   	2. 解析Source Entities
   	3. 解析Tentacles
   2. MQL运行时执行
   	1. json输入
   	2. 执行输出json
   	3. 输出的json反馈给Target
*/

type TentacleConfig struct {
	SourceEntity string
	PropertyKeys []string
}

type TQLConfig struct { // nolint
	TargetEntity   string
	SourceEntities []string
	Tentacles      []TentacleConfig
}

type TQL interface {
	Target() string
	Entities() []string
	Tentacles() []TentacleConfig
	Exec(map[string]constraint.Node) (map[string]constraint.Node, error)
}
