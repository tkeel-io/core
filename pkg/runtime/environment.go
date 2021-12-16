package runtime

import "github.com/tkeel-io/core/pkg/statem"

type Environment struct {
}

func NewEnv() *Environment {
	return &Environment{}
}

func (env *Environment) LoadMapper(descs []*statem.MapperDesc) error {
	return nil
}

func (env *Environment) OnMapperChanged(op string, desc *statem.MapperDesc) error {
	return nil
}
