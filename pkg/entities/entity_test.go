package entities

import (
	"context"
	"testing"

	ants "github.com/panjf2000/ants/v2"
)

func TestEntity(t *testing.T) {
	coroutinePool, err := ants.NewPool(500)
	if nil != err {
		panic(err)
	}

	tag := "test"
	mgr := NewEntityManager(context.Background(), coroutinePool)

	enty1, err1 := newEntity(context.Background(), mgr, "", "abcd", "tomas", &tag, 001)
	enty2, err2 := newEntity(context.Background(), mgr, "", "abcd", "tomas", &tag, 001)

	t.Log(enty1, enty2, err1, err2)
}
