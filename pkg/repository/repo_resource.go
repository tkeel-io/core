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

package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func (r *repo) PutCostumeResource(ctx context.Context, crd dao.Resource) error {
	return errors.Wrap(r.dao.PutResource(ctx, crd), "put mapper repository")
}

func (r *repo) GetCostumeResource(ctx context.Context, m dao.Resource) (dao.Resource, error) {
	mapper, err := r.dao.GetResource(ctx, m)
	return mapper, errors.Wrap(err, "get mapper repository")
}

func (r *repo) DelCostumeResource(ctx context.Context, m dao.Resource) error {
	return errors.Wrap(r.dao.DelResource(ctx, m), "del mapper repository")
}

//func (r *repo) DelResourceByEntity(ctx context.Context, m dao.Resource) error {
//	return errors.Wrap(r.dao.DelByEntity(ctx, m), "del mapper repository")
//}

func (r *repo) HasCostumeResource(ctx context.Context, m dao.Resource) (bool, error) {
	has, err := r.dao.HasResource(ctx, m)
	return has, errors.Wrap(err, "exists mapper repository")
}

func (r *repo) ListCostumeResource(ctx context.Context, rev int64, prefix string, decodeFunc dao.DecodeFunc) ([]dao.Resource, error) {
	mappers, err := r.dao.ListResource(ctx, rev, prefix, decodeFunc)
	return mappers, errors.Wrap(err, "list mapper repository")
}

func (r *repo) RangeCostumeResource(ctx context.Context, rev int64, prefix string, handler dao.RangeResourceFunc) {
	r.dao.RangeResource(ctx, rev, prefix, handler)
}

func (r *repo) WatchCostumeResource(ctx context.Context, rev int64, prefix string, handler dao.WatchResourceFunc) {
	go r.dao.WatchResource(ctx, rev, prefix, handler)
}
