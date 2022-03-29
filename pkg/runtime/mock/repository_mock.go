package mock

import (
	"context"

	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func NewRepo() repository.IRepository {
	return &repo{}
}

type repo struct {
}

func (r *repo) GetLastRevision(context.Context) int64                       { return 0 }
func (r *repo) PutEntity(context.Context, string, []byte) error             { return nil }
func (r *repo) GetEntity(context.Context, string) ([]byte, error)           { return nil, nil }
func (r *repo) DelEntity(context.Context, string) error                     { return nil }
func (r *repo) HasEntity(context.Context, string) (bool, error)             { return false, nil }
func (r *repo) PutMapper(context.Context, *dao.Mapper) error                { return nil }
func (r *repo) GetMapper(context.Context, *dao.Mapper) (*dao.Mapper, error) { return nil, nil }
func (r *repo) DelMapper(context.Context, *dao.Mapper) error                { return nil }
func (r *repo) DelMapperByEntity(context.Context, *dao.Mapper) error        { return nil }
func (r *repo) HasMapper(context.Context, *dao.Mapper) (bool, error)        { return false, nil }
func (r *repo) ListMapper(context.Context, int64, *dao.ListMapperReq) ([]dao.Mapper, error) {
	return nil, nil
}
func (r *repo) RangeMapper(ctx context.Context, rev int64, handler dao.MapperHandler)      {}
func (r *repo) WatchMapper(ctx context.Context, rev int64, handler dao.WatchMapperHandler) {}
