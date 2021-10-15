package source

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

type BaseSourceGenerator struct {
	SourceType Type
	Generator  OpenSourceHandler
}

func (bs *BaseSourceGenerator) Type() Type {
	return bs.SourceType
}
func (bs *BaseSourceGenerator) OpenSource(ctx context.Context, metadata Metadata, service common.Service) (ISource, error) {
	return bs.Generator(ctx, metadata, service)
}

var Generators = make(map[Type]Generator)

func Register(generator Generator) {
	Generators[generator.Type()] = generator
}

func OpenSource(ctx context.Context, metadata Metadata, service common.Service) (ISource, error) {
	generator, exists := Generators[metadata.Type]
	if !exists {
		return nil, errors.New("source generator not register")
	}

	s, err := generator.OpenSource(ctx, metadata, service)
	if err != nil {
		return nil, errors.Wrap(err, "open source err")
	}

	return s, nil
}
