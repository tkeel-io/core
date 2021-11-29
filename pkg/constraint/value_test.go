package constraint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigsFrom(t *testing.T) {
	now := time.Now().UnixNano() / 1e6
	data := map[string]interface{}{
		"id":             "property1",
		"type":           "int",
		"weight":         20,
		"enabled":        true,
		"enabled_search": true,
		"description":    "property instance.",
		"last_time":      now,
		"define":         map[string]interface{}{"max": 200},
	}

	cfg := Config{
		ID:                "property1",
		Type:              "int",
		Weight:            20,
		Enabled:           true,
		EnabledSearch:     true,
		EnabledTimeSeries: false,
		Description:       "property instance.",
		LastTime:          now,
		Define: map[string]interface{}{
			"max": 200,
		},
	}

	result, err := ParseConfigsFrom(data)
	assert.Nilf(t, err, "parse successful.")
	assert.Equal(t, cfg, result)
}
