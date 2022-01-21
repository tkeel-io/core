package repository

import (
	"context"

	"github.com/tkeel-io/core/pkg/repository/dao"
)

type IRepository interface {
	PutEntity(context.Context, *dao.Entity) error
	GetEntity(context.Context, *dao.Entity) (*dao.Entity, error)
	DelEntity(context.Context, *dao.Entity) error
	HasEntity(context.Context, *dao.Entity) (bool, error)
	PutMapper(context.Context, *dao.Mapper) error
	GetMapper(context.Context, *dao.Mapper) (*dao.Mapper, error)
	DelMapper(context.Context, *dao.Mapper) error
	HasMapper(context.Context, *dao.Mapper) (bool, error)
	RangeMapper(ctx context.Context, rev int64, handler dao.MapperHandler)
	WatchMapper(ctx context.Context, rev int64, handler dao.MapperHandler)
}
