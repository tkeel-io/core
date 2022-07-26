package repository

import (
	"context"
)

type IRepository interface {
	GetLastRevision(ctx context.Context) int64
	PutEntity(ctx context.Context, eid string, data []byte) error
	FlushEntity(ctx context.Context) error
	GetEntity(ctx context.Context, eid string) ([]byte, error)
	DelEntity(ctx context.Context, eid string) error
	HasEntity(ctx context.Context, eid string) (bool, error)
	PutExpression(ctx context.Context, expr *Expression) error
	GetExpression(ctx context.Context, expr *Expression) (*Expression, error)
	DelExpression(ctx context.Context, expr *Expression) error
	DelExprByEnity(ctx context.Context, expr *Expression) error
	HasExpression(ctx context.Context, expr *Expression) (bool, error)
	ListExpression(ctx context.Context, rev int64, req *ListExprReq) ([]*Expression, error)
	RangeExpression(ctx context.Context, rev int64, handler RangeExpressionFunc)
	WatchExpression(ctx context.Context, rev int64, handler WatchExpressionFunc)
	PutSubscription(ctx context.Context, expr *Subscription) error
	GetSubscription(ctx context.Context, expr *Subscription) (*Subscription, error)
	DelSubscription(ctx context.Context, expr *Subscription) error
	HasSubscription(ctx context.Context, expr *Subscription) (bool, error)
	RangeSubscription(ctx context.Context, rev int64, handler RangeSubscriptionFunc)
	WatchSubscription(ctx context.Context, rev int64, handler WatchSubscriptionFunc)
	PutSchema(ctx context.Context, expr *Schema) error
	GetSchema(ctx context.Context, expr *Schema) (*Schema, error)
	DelSchema(ctx context.Context, expr *Schema) error
	HasSchema(ctx context.Context, expr *Schema) (bool, error)
	RangeSchema(ctx context.Context, rev int64, handler RangeSchemaFunc)
	WatchSchema(ctx context.Context, rev int64, handler WatchSchemaFunc)
}
