package repository

import (
	"context"
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
}
