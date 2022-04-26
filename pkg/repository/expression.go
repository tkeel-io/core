package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

const (
	ExprTypeSub  = "sub"
	ExprTypeEval = "eval"
	ExprPrefix   = "/core/v1/expressions"
)

type ListExprReq struct {
	Owner    string
	EntityID string
}

var _ dao.Resource = (*Expression)(nil)

type Expression struct {
	// expression identifier.
	ID string
	// target path.
	Path string
	// expression name.
	Name string
	// expression type.
	Type string
	// expression owner.
	Owner string
	// entity id.
	EntityID string
	// expression.
	Expression string
	// description.
	Description string
}

func NewExpression(owner, entityID, name, path, expr, desc string) *Expression {
	escapePath := url.PathEscape(path)
	typ := ExprTypeEval
	if escapePath == "" {
		path = util.UUID("exprsub")
		escapePath = url.PathEscape(path)
		typ = ExprTypeSub
	}

	identifier := fmt.Sprintf("%s/%s/%s/%s",
		ExprPrefix, owner, entityID, escapePath)
	return &Expression{
		ID:          identifier,
		Name:        name,
		Path:        path,
		Type:        typ,
		Owner:       owner,
		EntityID:    entityID,
		Expression:  expr,
		Description: desc,
	}
}

func ListExpressionPrefix(Owner, EntityID string) string {
	keyString := fmt.Sprintf("%s/%s/%s",
		ExprPrefix, Owner, EntityID)
	return keyString
}

func (e *Expression) EncodeKey() ([]byte, error) {
	escapePath := url.PathEscape(e.Path)
	keyString := fmt.Sprintf("%s/%s/%s/%s",
		ExprPrefix, e.Owner, e.EntityID, escapePath)
	return []byte(keyString), nil
}

func (e *Expression) Encode() ([]byte, error) {
	bytes, err := json.Marshal(e)
	return bytes, errors.Wrap(err, "encode Expression")
}

func (e *Expression) Decode(bytes []byte) error {
	err := json.Unmarshal(bytes, e)
	return errors.Wrap(err, "decode Expression")
}

func (e *Expression) Prefix() string {
	return ListExpressionPrefix(e.Owner, e.EntityID)
}

func (r *repo) PutExpression(ctx context.Context, expr Expression) error {
	err := r.dao.PutResource(ctx, &expr)
	return errors.Wrap(err, "put expression repository")
}

func (r *repo) GetExpression(ctx context.Context, expr Expression) (Expression, error) {
	_, err := r.dao.GetResource(ctx, &expr)
	return expr, errors.Wrap(err, "get expression repository")
}

func (r *repo) DelExpression(ctx context.Context, expr Expression) error {
	err := r.dao.DelResource(ctx, &expr)
	return errors.Wrap(err, "del expression repository")
}

func (r *repo) DelExprByEnity(ctx context.Context, expr Expression) error {
	// construct prefix key.
	prefix := expr.Prefix()
	err := r.dao.DelResources(ctx, prefix)
	return errors.Wrap(err, "del expressions repository")
}

func (r *repo) HasExpression(ctx context.Context, expr Expression) (bool, error) {
	has, err := r.dao.HasResource(ctx, &expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListExpression(ctx context.Context, rev int64, req *ListExprReq) ([]*Expression, error) {
	// construct prefix.
	prefix := ListExpressionPrefix(req.Owner, req.EntityID)
	ress, err := r.dao.ListResource(ctx, rev, prefix,
		func(raw []byte) (dao.Resource, error) {
			var res Expression // escape.
			err := res.Decode(raw)
			return &res, errors.Wrap(err, "decode expression")
		})

	var exprs []*Expression
	for index := range ress {
		if expr, ok := ress[index].(*Expression); ok {
			exprs = append(exprs, expr)
			continue
		}
		// panic.
	}
	return exprs, errors.Wrap(err, "list expression repository")
}

func (r *repo) RangeExpression(ctx context.Context, rev int64, handler RangeExpressionFunc) {
	r.dao.RangeResource(ctx, rev, ExprPrefix, func(kvs []*mvccpb.KeyValue) {
		var exprs []*Expression
		for index := range kvs {
			var expr Expression
			err := expr.Decode(kvs[index].Value)
			if nil != err {
				log.L().Error("")
				continue
			}
			exprs = append(exprs, &expr)
		}
		handler(exprs)
	})
}

func (r *repo) WatchExpression(ctx context.Context, rev int64, handler WatchExpressionFunc) {
	r.dao.WatchResource(ctx, rev, ExprPrefix, func(et dao.EnventType, kv *mvccpb.KeyValue) {
		var expr Expression
		err := expr.Decode(kv.Value)
		if nil != err {
			log.L().Error("")
		}
		handler(et, expr)
	})
}

type RangeExpressionFunc func([]*Expression)
type WatchExpressionFunc func(dao.EnventType, Expression)
