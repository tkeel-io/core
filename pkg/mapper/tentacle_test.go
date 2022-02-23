package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTentacle(t *testing.T) {
	tentacle := NewTentacle(TentacleTypeEntity, "device123", []WatchKey{{
		EntityID:    "device234",
		PropertyKey: "temp",
	}}, 0)

	assert.Equal(t, "device123", tentacle.TargetID())
}
