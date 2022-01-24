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
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
)

func TestNewStatem(t *testing.T) {
	stateManager := NewManagerMock()

	base := dao.Entity{
		ID:         "device123",
		Type:       "DEVICE",
		Owner:      "admin",
		Source:     "dm",
		Version:    0,
		LastTime:   util.UnixMilli(),
		Properties: map[string]constraint.Node{"temp": constraint.NewNode(25)},
		ConfigFile: nil,
	}

	sm, err := NewState(context.Background(), stateManager, &base, nil)
	assert.Nil(t, err)
	assert.Equal(t, "device123", sm.GetID())
}
