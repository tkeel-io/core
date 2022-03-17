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
func (r *repo) PutEntity(context.Context, *dao.Entity) error                { return nil }
func (r *repo) GetEntity(context.Context, *dao.Entity) (*dao.Entity, error) { return nil, nil }
func (r *repo) DelEntity(context.Context, *dao.Entity) error                { return nil }
func (r *repo) HasEntity(context.Context, *dao.Entity) (bool, error)        { return false, nil }
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
func (r *repo) PutQueue(context.Context, *dao.Queue) error                                 { return nil }
func (r *repo) GetQueue(context.Context, *dao.Queue) (*dao.Queue, error)                   { return nil, nil }
func (r *repo) DelQueue(context.Context, *dao.Queue) error                                 { return nil }
func (r *repo) HasQueue(context.Context, *dao.Queue) (bool, error)                         { return false, nil }
func (r *repo) RangeQueue(ctx context.Context, rev int64, handler dao.QueueHandler)        {}
func (r *repo) WatchQueue(ctx context.Context, rev int64, handler dao.WatchQueueHandler)   {}
