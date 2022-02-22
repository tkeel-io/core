package dao

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tdtl"
)

func TestEntity(t *testing.T) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]tdtl.Node{
			"temp": tdtl.FloatNode(123.3),
		},
	}

	assert.Equal(t, "device123", en.ID)
	assert.Equal(t, "BASIC", en.Type)
	assert.Equal(t, "admin", en.Owner)
	assert.Equal(t, "dm", en.Source)
}

func TestEntity_Copy(t *testing.T) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]tdtl.Node{
			"temp": tdtl.FloatNode(123.3),
		},
	}

	cp := en.Copy()
	assert.Equal(t, en.ID, cp.ID)
	assert.Equal(t, en.Type, cp.Type)
	assert.Equal(t, en.Owner, cp.Owner)
	assert.Equal(t, en.Source, cp.Source)
}

func TestEntity_Basic(t *testing.T) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]tdtl.Node{
			"temp": tdtl.FloatNode(123.3),
		},
	}

	cp := en.Basic()
	assert.Equal(t, en.ID, cp.ID)
	assert.Equal(t, en.Type, cp.Type)
	assert.Equal(t, en.Owner, cp.Owner)
	assert.Equal(t, en.Source, cp.Source)
}

func BenchmarkCopy(b *testing.B) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]tdtl.Node{
			"temp":     tdtl.FloatNode(123.3),
			"metrics0": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics1": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics2": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics3": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics4": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics5": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics6": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics7": tdtl.JSONNode(`{"cpu_used": 0.2, "mem_used": 0.7}`),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en.Copy()
	}
}
