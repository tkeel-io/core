package source

import (
	"context"
	"errors"

	"github.com/dapr/go-sdk/service/common"
)

type BaseSourceGenerator struct {
	SourceType SourceType
	Generator  OpenSourceHandler
}

func (bs *BaseSourceGenerator) Type() SourceType {
	return bs.SourceType
}
func (bs *BaseSourceGenerator) OpenSource(ctx context.Context, metadata Metadata, service common.Service) (ISource, error) {
	return bs.Generator(ctx, metadata, service)
}

var SourceGenerators = make(map[SourceType]SourceGenerator)

func Register(generator SourceGenerator) {
	SourceGenerators[generator.Type()] = generator
}

func OpenSource(ctx context.Context, metadata Metadata, service common.Service) (ISource, error) {
	Generator, exists := SourceGenerators[metadata.Type]
	if !exists {
		return nil, errors.New("source Generator not register.")
	}

	return Generator.OpenSource(ctx, metadata, service)
}
