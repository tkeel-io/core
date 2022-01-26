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

package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/runtime/message"
)

func Test_newMailbox(t *testing.T) {
	mb := newMailbox(5)

	assert.Equal(t, 5, mb.Capcity())
}

func Test_Resize(t *testing.T) {
	mb := newMailbox(5)
	assert.Equal(t, 5, mb.Capcity())

	mb.Resize(20)
	assert.Equal(t, 20, mb.Capcity())
	err := mb.Resize(10)
	assert.NotNil(t, err)
	assert.Equal(t, 20, mb.Capcity())
	mb.Resize(30)
	assert.Equal(t, 30, mb.Capcity())
}

func Test_Put(t *testing.T) {
	mb := newMailbox(5)

	ctx := message.New(context.TODO())
	mb.Put(ctx)
	mb.Put(ctx)
	mb.Put(ctx)
	assert.Equal(t, 3, mb.Size())
	assert.Equal(t, 5, mb.Capcity())
}

func Test_Get(t *testing.T) {
	mb := newMailbox(5)
	mb.Put(message.New(context.TODO()))
	assert.NotNil(t, mb.Get())
}
