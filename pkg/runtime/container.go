package runtime

import "github.com/tkeel-io/core/pkg/entities"

type Container struct {
	size     int
	entities map[string]entities.EntityOp
}

func NewContainer() *Container {
	return &Container{
		size:     0,
		entities: make(map[string]entities.EntityOp),
	}
}

func (c *Container) Size() int {
	return c.size
}

func (c *Container) Get(eid string) entities.EntityOp {
	return c.entities[eid]
}

func (c *Container) Set(e entities.EntityOp) error {
	c.entities[e.GetID()] = e
	return nil
}
