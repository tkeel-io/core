package entities

import (
	"context"
	"testing"
)

func TestEntity_GetEntity(t *testing.T) {
	m := NewEntityManager(context.Background(), nil)
	m.getEntity("id")
}
