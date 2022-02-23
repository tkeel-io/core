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

package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/atomic"
)

type SearchService struct {
	pb.UnimplementedSearchServer

	inited       *atomic.Bool
	searchClient pb.SearchHTTPServer
}

func NewSearchService() *SearchService {
	return &SearchService{
		inited: atomic.NewBool(false),
	}
}

func (s *SearchService) Init(searchClient pb.SearchHTTPServer) {
	s.searchClient = searchClient
	s.inited.Store(true)
}

func (s *SearchService) Index(ctx context.Context, req *pb.IndexObject) (*pb.IndexResponse, error) {
	out, err := s.searchClient.Index(ctx, req)
	if err != nil {
		return out, errors.Wrap(err, "index failed")
	}
	return out, nil
}
func (s *SearchService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if !s.inited.Load() {
		log.Warn("service not ready")
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	out, err := s.searchClient.Search(ctx, req)
	if err != nil {
		return out, errors.Wrap(err, "search failed")
	}
	return out, nil
}
