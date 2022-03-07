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
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContraint1(t *testing.T) {
	cfg := Config{
		ID:                "property1",
		Type:              "int",
		Weight:            20,
		Enabled:           true,
		EnabledSearch:     true,
		EnabledTimeSeries: false,
		Description:       "property instance.",
		LastTime:          time.Now().UnixNano() / 1e6,
		Define: map[string]interface{}{
			"max": 200,
		},
	}

	ct := NewConstraintsFrom(cfg)

	assert.Equal(t, ct.ID, cfg.ID)
	assert.Equal(t, ct.Type, cfg.Type)
	assert.Equal(t, len(ct.Operators), 1)
	assert.Equal(t, ct.EnableFlag.Enabled(EnabledFlagSelf), true)
	assert.Equal(t, ct.EnableFlag.Enabled(EnabledFlagSearch), true)
	assert.Equal(t, ct.Operators, []Operator{{Callback: "max", Condition: 200}})
	assert.Equal(t, ct.GenEnabledIndexes(EnabledFlagSearch), []string{"property1"})
}

func TestContraint2(t *testing.T) {
	cfg := Config{
		ID:                "property2",
		Type:              "struct",
		Weight:            20,
		Enabled:           true,
		EnabledSearch:     true,
		EnabledTimeSeries: false,
		Description:       "property instance.",
		LastTime:          time.Now().UnixNano() / 1e6,
		Define: map[string]interface{}{
			"fields": map[string]Config{
				"property2-1": {
					ID:                "property2-1",
					Type:              "int",
					Weight:            20,
					Enabled:           true,
					EnabledSearch:     true,
					EnabledTimeSeries: false,
					Description:       "property instance.",
					LastTime:          time.Now().UnixNano() / 1e6,
					Define: map[string]interface{}{
						"max": 200,
					},
				},
				"property2-2": {
					ID:                "property2-2",
					Type:              "string",
					Weight:            20,
					Enabled:           true,
					EnabledSearch:     true,
					EnabledTimeSeries: false,
					Description:       "property instance.",
					LastTime:          time.Now().UnixNano() / 1e6,
					Define: map[string]interface{}{
						"size": 256,
					},
				},
			},
		},
	}

	ct := NewConstraintsFrom(cfg)

	assert.Equal(t, ct.ID, cfg.ID)
	assert.Equal(t, ct.Type, cfg.Type)
	assert.Equal(t, len(ct.Operators), 0)
	assert.Equal(t, ct.EnableFlag.Enabled(EnabledFlagSelf), true)
	assert.Equal(t, ct.EnableFlag.Enabled(EnabledFlagSearch), true)
	assert.Equal(t, 2, len(ct.ChildNodes))

	var ret sort.StringSlice = ct.GenEnabledIndexes(EnabledFlagSearch)
	sort.Sort(ret)
	assert.Equal(t, []string(ret), []string{"property2", "property2.property2-1", "property2.property2-2"})
}

func TestFormatPropertyKey(t *testing.T) {
	pathes := []string{"a.b.c.d"}
	FormatPropertyKey(pathes)
	assert.Equal(t, []string{"a.define.fields.b.define.fields.c.define.fields.d"}, pathes)
}
