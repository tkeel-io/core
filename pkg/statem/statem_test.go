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

package statem

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/constraint"
)

func TestNewStatem(t *testing.T) {
	stateManager := NewStateManagerMock()

	base := Base{
		ID:           "device123",
		Type:         "DEVICE",
		Owner:        "admin",
		Source:       "dm",
		Version:      0,
		LastTime:     time.Now().UnixMilli(),
		Mappers:      []MapperDesc{{Name: "mapper123", TQLString: "insert into device123 select device234.temp as temp"}},
		KValues:      map[string]constraint.Node{"temp": constraint.NewNode(25)},
		ConfigsBytes: nil,
	}

	sm, err := NewState(context.Background(), stateManager, &base, nil)
	assert.Nil(t, err)
	assert.Equal(t, "device123", sm.GetID())
	assert.Equal(t, "DEVICE", sm.GetBase().Type)
	assert.Equal(t, "admin", sm.GetBase().Owner)
	assert.Equal(t, SMStatusActive, sm.GetStatus())
}
