package constraint

import (
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
	assert.Equal(t, ct.GenSearchIndex(), []string{"property1"})
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
			"fields": []Config{
				{
					ID:                "property2.1",
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
				{
					ID:                "property2.2",
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
	assert.Equal(t, len(ct.ChildNodes), 2)
	assert.Equal(t, ct.GenSearchIndex(), []string{"property2", "property2.property2.1", "property2.property2.2"})
}
