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
	res, loaded := m.LoadOrStore("xxx", struct{}{})
	t.Log(res, loaded)
}
