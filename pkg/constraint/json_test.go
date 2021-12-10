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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	valInt := 23
	valFloat := 55.8

	assert.Equal(t, NewNode(true), BoolNode(true), "BoolNode<true>.")
	assert.Equal(t, NewNode(false), BoolNode(false), "BoolNode<false>.")
	assert.Equal(t, NewNode(int(1)), IntNode(1), "IntNode.")
	assert.Equal(t, NewNode(1.2), FloatNode(1.2), "FloatNode.")
	assert.Equal(t, NewNode(-22.1).To(String), StringNode("-22.1"), "StringNode.")
	assert.Equal(t, NewNode(&valInt), IntNode(valInt), "IntNode PTR.")
	assert.Equal(t, NewNode(&valFloat), FloatNode(valFloat), "FloatNode PTR.")
	assert.Equal(t, NewNode([]byte("test bytes.")), JSONNode("test bytes."), "RawNode PTR.")

	t.Log(NewNode(-22.1).To(String).String())

	time.Sleep(time.Second)
}
