package entities

import (
	"context"
	"fmt"
	"testing"
)

func TestEntity(t *testing.T) {

	mgr := NewManager()

	enty1, err := NewEntity(context.Background(), mgr, "", "abcd", "tomas", "test", 001)
	enty2, err := NewEntity(context.Background(), mgr, "", "abcd", "tomas", "test", 001)

	fmt.Println(enty1, enty2, err)
}
