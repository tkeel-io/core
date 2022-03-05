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

package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/core/api/core/v1"
)

func TestNewEntity(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)
	t.Log(string(en.Raw()))
}

func TestEntity_Handle(t *testing.T) {
	en, err := NewEntity("en-123", []byte(`{"properties": {"temp": 20}}`))
	assert.Nil(t, err)

	tests := map[string]v1.PatchEvent{
		"patch1": &v1.ProtoEvent{
			Id: "ev-1",
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{
						{
							Path:     "properties.temp",
							Value:    []byte("50"),
							Operator: string(OpReplace),
						},
						{
							Path:     "properties.metrics.cpu_used",
							Value:    []byte("0.78"),
							Operator: string(OpReplace),
						},
						{
							Path:     "properties.metrics.mem_used",
							Value:    []byte("0.28"),
							Operator: string(OpReplace),
						},
						{
							Path:     "properties.metrics.interfaces",
							Value:    []byte("0.28"),
							Operator: string(OpAdd),
						},
						{
							Path:     "properties.temp",
							Operator: string(OpRemove),
						},
						{
							Path:     "properties.metrics",
							Value:    []byte(`{"temp": 209}`),
							Operator: string(OpMerge),
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := en.Handle(context.TODO(), test)
			assert.Nil(t, err)
			t.Log("result", result)
		})
	}

	t.Log(string(en.Raw()))
}
