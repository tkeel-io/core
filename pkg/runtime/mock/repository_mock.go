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

func (r *repo) PutExpression(ctx context.Context, expr dao.Expression) error { return nil }
func (r *repo) GetExpression(ctx context.Context, expr dao.Expression) (dao.Expression, error) {
	return dao.Expression{}, nil
}
func (r *repo) DelExpression(ctx context.Context, expr dao.Expression) error  { return nil }
func (r *repo) DelExprByEnity(ctx context.Context, expr dao.Expression) error { return nil }
func (r *repo) HasExpression(ctx context.Context, expr dao.Expression) (bool, error) {
	return false, nil
}
func (r *repo) ListExpression(ctx context.Context, rev int64, req *dao.ListExprReq) ([]dao.Expression, error) {
	return nil, nil
}
func (r *repo) RangeExpression(ctx context.Context, rev int64, handler dao.ExpressionFunc)      {}
func (r *repo) WatchExpression(ctx context.Context, rev int64, handler dao.WatchExpressionFunc) {}
