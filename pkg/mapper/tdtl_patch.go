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

package mapper

import (
	"strings"

	"github.com/tkeel-io/tdtl"
)

// ------------------------.
type TentacleConfig struct {
	SourceEntity string
	PropertyKeys []string
}

type TQLConfig struct {
	TargetEntity   string
	SourceEntities []string
	Tentacles      []TentacleConfig
}

func tentacles(tl tdtl.TDTL) []TentacleConfig {
	tentacleCfgs := make([]TentacleConfig, 0, len(tl.Entities()))
	for entityID, keys := range tl.Entities() {
		tentacleCfg := TentacleConfig{SourceEntity: entityID}
		for index := range keys {
			arr := strings.SplitN(keys[index], ".", 2)
			tentacleCfg.PropertyKeys = append(tentacleCfg.PropertyKeys, arr[1])
		}
		tentacleCfgs = append(tentacleCfgs, tentacleCfg)
	}
	return tentacleCfgs
}
