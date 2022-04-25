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

package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/mapper"
)

func TestEntity_GetEntity(t *testing.T) {
	// NewAPIManager(context.Background(), nil, nil).
}

func Test_checkTQL(t *testing.T) {
	mappers := []struct {
		Name   string
		TQL    string
		Expect string
	}{{
		Name:   "test1",
		TQL:    "insert into device123 select device234.metrics as metrics, device234.metrics.cpu as cpu",
		Expect: "insert into device123 select device234.properties.metrics as properties.metrics, device234.properties.metrics.cpu as properties.cpu",
	}, {
		Name:   "test2",
		TQL:    "insert into sub123 select device123.*",
		Expect: "insert into sub123 select device123.*",
	}}

	for index := range mappers {
		t.Run(mappers[index].Name, func(t *testing.T) {
			mp := &mapper.Mapper{Name: mappers[index].Name, TQL: mappers[index].TQL}
			err := checkMapper(mp)
			assert.Nil(t, err)
			assert.Equal(t, mappers[index].Expect, mp.TQL)
		})
	}
}
