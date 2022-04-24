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
	PutExpression(ctx context.Context, expr Expression) error
	GetExpression(ctx context.Context, expr Expression) (Expression, error)
	DelExpression(ctx context.Context, expr Expression) error
	DelExprByEnity(ctx context.Context, expr Expression) error
	HasExpression(ctx context.Context, expr Expression) (bool, error)
	ListExpression(ctx context.Context, rev int64, req *ListExprReq) ([]*Expression, error)
	RangeExpression(ctx context.Context, rev int64, handler RangeExpressionFunc)
	WatchExpression(ctx context.Context, rev int64, handler WatchExpressionFunc)
	PutCostumeResource(ctx context.Context, expr dao.Resource) error
	GetCostumeResource(ctx context.Context, expr dao.Resource) (dao.Resource, error)
	DelCostumeResource(ctx context.Context, expr dao.Resource) error
	HasCostumeResource(ctx context.Context, expr dao.Resource) (bool, error)
	ListCostumeResource(ctx context.Context, rev int64, prefix string, decodeFunc dao.DecodeFunc) ([]dao.Resource, error)
	RangeCostumeResource(ctx context.Context, rev int64, prefix string, handler dao.RangeResourceFunc)
	WatchCostumeResource(ctx context.Context, rev int64, prefix string, handler dao.WatchResourceFunc)
}
