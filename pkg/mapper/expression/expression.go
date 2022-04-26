package expression

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/tdtl"
)

type IExpression interface {
	Eval(context.Context, map[string]tdtl.Node) (tdtl.Node, error)
	Entities() map[string][]string
}

func Validate(expr repository.Expression) error {
	// check path.

	// check expression.
	_, err := NewExpr(expr.Expression, nil)
	return errors.Wrap(err, "invalid expression")
}

type Expr struct {
	exprIns tdtl.Expression
}

func NewExpr(expression string, extFuncs map[string]tdtl.ContextFunc) (IExpression, error) {
	exprIns, err := tdtl.NewExpr(expression, extFuncs)
	return &Expr{exprIns: exprIns}, errors.Wrap(err, "new expression evaler")
}

func (e *Expr) Eval(ctx context.Context, in map[string]tdtl.Node) (tdtl.Node, error) {
	result := e.exprIns.Eval(in)
	return result, errors.Wrap(result.Error(), "eval expression")
}

func (e *Expr) Entities() map[string][]string {
	return e.exprIns.Sources()
}
