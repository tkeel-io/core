package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	_ "github.com/tkeel-io/core/pkg/resource/store/dapr"
	_ "github.com/tkeel-io/core/pkg/resource/store/noop"
	"github.com/tkeel-io/tdtl"
)

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
			"metrics0": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics1": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics2": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics3": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics4": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics5": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics6": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
			"metrics7": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.7}`),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en.Copy()
	}
}

func Test_PutEntity(t *testing.T) {
	err := repoIns.PutEntity(context.Background(), "device123", []byte(`{}`))
	assert.Nil(t, err)
}

func Test_GetEntity(t *testing.T) {
	_, err := repoIns.GetEntity(context.Background(), "device123")
	assert.Equal(t, true, errors.Is(err, xerrors.ErrResourceNotFound))
}

func Test_DelEntity(t *testing.T) {
	err := repoIns.DelEntity(context.Background(), "device123")
	assert.Nil(t, err)
}

func Test_HasEntity(t *testing.T) {
	has, err := repoIns.HasEntity(context.Background(), "device123")
	assert.Nil(t, err)
	assert.Equal(t, false, has)
}
