package repository

import (
	"context"

	"github.com/tkeel-io/core/pkg/repository/dao"
)

type IRepository interface {
	GetLastRevision(ctx context.Context) int64
	PutEntity(ctx context.Context, eid string, data []byte) error
	GetEntity(ctx context.Context, eid string) ([]byte, error)
	DelEntity(ctx context.Context, eid string) error
	HasEntity(ctx context.Context, eid string) (bool, error)
	PutMapper(ctx context.Context, mp *dao.Mapper) error
	GetMapper(ctx context.Context, mp *dao.Mapper) (*dao.Mapper, error)
	DelMapper(ctx context.Context, mp *dao.Mapper) error
	DelMapperByEntity(ctx context.Context, mp *dao.Mapper) error
	HasMapper(ctx context.Context, mp *dao.Mapper) (bool, error)
	ListMapper(ctx context.Context, rev int64, req *dao.ListMapperReq) ([]dao.Mapper, error)
	RangeMapper(ctx context.Context, rev int64, handler dao.MapperHandler)
	WatchMapper(ctx context.Context, rev int64, handler dao.WatchMapperHandler)
	PutExpression(ctx context.Context, expr dao.Expression) error
	GetExpression(ctx context.Context, expr dao.Expression) (dao.Expression, error)
	DelExpression(ctx context.Context, expr dao.Expression) error
	DelExprByEnity(ctx context.Context, expr dao.Expression) error
	HasExpression(ctx context.Context, expr dao.Expression) (bool, error)
	ListExpression(ctx context.Context, rev int64, req *dao.ListExprReq) ([]dao.Expression, error)
	RangeExpression(ctx context.Context, rev int64, handler dao.ExpressionFunc)
	WatchExpression(ctx context.Context, rev int64, handler dao.WatchExpressionFunc)
	PutCostumeResource(ctx context.Context, expr dao.CostumeResource) error
	GetCostumeResource(ctx context.Context, expr dao.CostumeResource) (dao.CostumeResource, error)
	DelCostumeResource(ctx context.Context, expr dao.CostumeResource) error
	HasCostumeResource(ctx context.Context, expr dao.CostumeResource) (bool, error)
	ListCostumeResource(ctx context.Context, rev int64, req dao.CostumeResourceReq) ([]dao.CostumeResource, error)
	RangeCostumeResource(ctx context.Context, rev int64, handler dao.ListCostumeResourceFunc)
	WatchCostumeResource(ctx context.Context, rev int64, handler dao.WatchCostumeResourceFunc)
}
