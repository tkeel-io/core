package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/tdtl"
)

var repoIns IRepository
var testReady bool

func TestMain(m *testing.M) {
	// create dao instance.
	daoIns, err := dao.New(context.Background(), config.Metadata{
		Name: "dapr",
		Properties: []config.Pair{
			{
				Key:   "store_name",
				Value: "core-state",
			},
		},
	}, config.EtcdConfig{
		Endpoints:   []string{"heep://localhost:2379"},
		DialTimeout: 3,
	})

	if nil != err {
		os.Exit(1)
	}
	testReady = true
	repoIns = New(daoIns)
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

}

func Test_GetEntity(t *testing.T) {

}

func Test_DelEntity(t *testing.T) {

}

func Test_HasEntity(t *testing.T) {

}
