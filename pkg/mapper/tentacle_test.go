package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTentacle(t *testing.T) {
	tentacle := NewTentacle(TentacleTypeEntity, "device123", []WatchKey{{
		EntityId:    "device234",
		PropertyKey: "temp",
	}})

	assert.Equal(t, "device123", tentacle.TargetID())
}
