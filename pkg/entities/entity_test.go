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

func TestGetProperties(t *testing.T) {
	coroutinePool, err := ants.NewPool(500)
	if nil != err {
		panic(err)
	}

	mgr := NewEntityManager(context.Background(), coroutinePool)

	entity, err := newEntity(context.Background(), mgr, "", "abcd", "tomas", nil, 001)
	if nil != err {
		panic(err)
	}

	entity.SetProperties(map[string]interface{}{
		"temp1":   15,
		"temp2":   23,
		"light":   555,
		"say":     "hello",
		"friends": []string{"tom", "tony"},
		"user": map[string]interface{}{
			"name": "john",
			"age":  20,
		},
	})

	props := entity.GetAllProperties()
	t.Log(props)

	// delete some field.
	delete(props, "light")

	props1 := entity.GetAllProperties()
	t.Log(props1)
}
