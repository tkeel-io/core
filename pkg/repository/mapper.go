package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func (r *repo) PutMapper(ctx context.Context, m *dao.Mapper) error {
	return errors.Wrap(r.dao.PutMapper(ctx, m), "put mapper repository")
}

func (r *repo) GetMapper(ctx context.Context, m *dao.Mapper) (*dao.Mapper, error) {
	mapper, err := r.dao.GetMapper(ctx, m)
	return mapper, errors.Wrap(err, "get mapper repository")
}

func (r *repo) DelMapper(ctx context.Context, m *dao.Mapper) error {
	return errors.Wrap(r.dao.DelMapper(ctx, m), "put mapper repository")
}

func (r *repo) HasMapper(ctx context.Context, m *dao.Mapper) (bool, error) {
	has, err := r.dao.HasMapper(ctx, m)
	return has, errors.Wrap(err, "exists mapper repository")
}

func (r *repo) ListMapper(ctx context.Context, rev int64, req *dao.ListMapperReq) ([]dao.Mapper, error) {
	mappers, err := r.dao.ListMapper(ctx, rev, req)
	return mappers, errors.Wrap(err, "list mapper repository")
}

func (r *repo) RangeMapper(ctx context.Context, rev int64, handler dao.MapperHandler) {
	r.dao.RangeMapper(ctx, rev, handler)
}

func (r *repo) WatchMapper(ctx context.Context, rev int64, handler dao.WatchMapperHandler) {
	go r.dao.WatchMapper(ctx, rev, handler)
}
