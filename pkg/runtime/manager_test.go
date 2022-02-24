package runtime

import (
	"sync"
	"testing"
)

func Test_getMachiner(t *testing.T) {
	// pool, _ := ants.NewPool(200)
	// stateManager, _ := NewManager(context.Background(), pool, nil)
}

func TestSyncMap(t *testing.T) {
	m := sync.Map{}
	m.Store("xxx", "xxx")
	res, loaded := m.LoadOrStore("xxx", "xxxxxxx")
	t.Log(res, loaded)
}
