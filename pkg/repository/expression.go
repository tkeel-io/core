package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func (r *repo) PutExpression(ctx context.Context, expr dao.Expression) error {
	return errors.Wrap(r.dao.PutExpression(ctx, expr), "put expression repository")
}

func (r *repo) GetExpression(ctx context.Context, expr dao.Expression) (dao.Expression, error) {
	expression, err := r.dao.GetExpression(ctx, expr)
	return expression, errors.Wrap(err, "get expression repository")
}

func (r *repo) DelExpression(ctx context.Context, expr dao.Expression) error {
	return errors.Wrap(r.dao.DelExpression(ctx, expr), "del expression repository")
}

func (r *repo) DelExprByEnity(ctx context.Context, expr dao.Expression) error {
	return errors.Wrap(r.dao.DelByEnity(ctx, expr), "del expression repository")
}

func (r *repo) HasExpression(ctx context.Context, expr dao.Expression) (bool, error) {
	has, err := r.dao.HasExpression(ctx, expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListExpression(ctx context.Context, rev int64, req *dao.ListExprReq) ([]dao.Expression, error) {
	exprs, err := r.dao.ListExpression(ctx, rev, req)
	return exprs, errors.Wrap(err, "list expression repository")
}

func (r *repo) RangeExpression(ctx context.Context, rev int64, handler dao.ExpressionFunc) {
	r.dao.RangeExpression(ctx, rev, handler)
}

func (r *repo) WatchExpression(ctx context.Context, rev int64, handler dao.WatchExpressionFunc) {
	go r.dao.WatchExpression(ctx, rev, handler)
}
