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

package dao

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/tkeel-io/core/pkg/constraint"
// )

// func TestDao(t *testing.T) {
// 	d := New(context.Background(), "core-entity", &storeMock{})
// 	assert.NotNil(t, d, "result not nil")
// }

// func TestDao_PutEntity(t *testing.T) {
// 	d := New(context.Background(), "core-entity", &storeMock{})
// 	assert.NotNil(t, d, "result not nil")
// 	err := d.PutEntity(context.TODO(), &Entity{
// 		ID:         "device123",
// 		Type:       "DEVICE",
// 		Owner:      "admin",
// 		Source:     "dm",
// 		Version:    0,
// 		Properties: map[string]constraint.Node{"temp": constraint.NewNode(25)},
// 	})
// 	assert.Nil(t, err, "nil error")
// }

// func TestDao_GetEntity(t *testing.T) {
// 	d := New(context.Background(), "core-entity", &storeMock{})
// 	assert.NotNil(t, d, "result not nil")
// 	en, err := d.GetEntity(context.TODO(), "device123")
// 	assert.Nil(t, err, "nil error")
// 	assert.Equal(t, "device123", en.ID)
// }

// func TestDao_DelEntity(t *testing.T) {
// 	d := New(context.Background(), "core-entity", &storeMock{})
// 	assert.NotNil(t, d, "result not nil")
// 	err := d.DelEntity(context.TODO(), "device123")
// 	assert.Nil(t, err, "nil error")
// }
