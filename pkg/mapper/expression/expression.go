package expression

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/tdtl"
)

type IExpression interface {
	ID() string
	Eval(context.Context, map[string]tdtl.Node) (map[string]tdtl.Node, error)
}

func Validate(expr dao.Expression) error {
	return nil
}

type Expr struct {
	id      string
	evalIns tdtl.TDTL
}

func NewExpr(expr dao.Expression) (IExpression, error) {
	exprIns, err := tdtl.NewExpr(expr.Expression, nil)
	return &Expr{id: expr.ID, evalIns: exprIns}, errors.Wrap(err, "new expression evaler")
}

func (e *Expr) ID() string {
	return e.id
}

func (e *Expr) Eval(ctx context.Context, in map[string]tdtl.Node) (map[string]tdtl.Node, error) {
	return nil, nil
}
