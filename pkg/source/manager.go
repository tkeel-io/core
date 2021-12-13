/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
