package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

const (
	ExprTypeSub   = "sub"
	ExprTypeEval  = "eval"
	fmtExprPrefix = "/core/v1/expressions"
)

type ListExprReq struct {
	Owner    string
	EntityID string
}

type defaultExprCodec struct{}

func (ec *defaultExprCodec) Key() dao.Codec {
	return &defaultExprKeyCodec{}
}

func (ec *defaultExprCodec) Value() dao.Codec {
	return &defaultExprValueCodec{}
}

type defaultExprKeyCodec struct{}
type defaultExprValueCodec struct{}

func encodeKey(entityID, owner, path string) string {
	escapePath := url.PathEscape(path)
	keyString := fmt.Sprintf("%s/%s/%s/%s",
		fmtExprPrefix, owner, entityID, escapePath)
	return keyString
}

func encodePrefix(entityID, owner string) string {
	keyString := fmt.Sprintf("%s/%s/%s",
		fmtExprPrefix, owner, entityID)
	return keyString
}

func (dec *defaultExprKeyCodec) Encode(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case *Expression:
		keyString := encodeKey(val.EntityID, val.Owner, val.Path)
		return []byte(keyString), nil
	default:
		return nil, xerrors.ErrInternal
	}
}

func (dec *defaultExprKeyCodec) Decode(raw []byte, v interface{}) error {
	panic("never use")
}

func (dec *defaultExprValueCodec) Encode(v interface{}) ([]byte, error) {
	bytes, err := json.Marshal(v)
	return bytes, errors.Wrap(err, "encode Expression")
}

func (dec *defaultExprValueCodec) Decode(raw []byte, v interface{}) error {
	err := json.Unmarshal(raw, v)
	return errors.Wrap(err, "decode Expression")
}

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
		fmtExprPrefix, owner, entityID, escapePath)
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

func (e *Expression) Codec() dao.KVCodec {
	return &defaultExprCodec{}
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
	prefix := encodePrefix(expr.EntityID, expr.Owner)
	err := r.dao.DelResources(ctx, prefix)
	return errors.Wrap(err, "del expressions repository")
}

func (r *repo) HasExpression(ctx context.Context, expr Expression) (bool, error) {
	has, err := r.dao.HasResource(ctx, &expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListExpression(ctx context.Context, rev int64, req *ListExprReq) ([]*Expression, error) {
	// construct prefix.
	prefix := encodePrefix(req.EntityID, req.Owner)
	ress, err := r.dao.ListResource(ctx, rev, prefix,
		func(raw []byte) (dao.Resource, error) {
			var res Expression // escape.
			valCodec := &defaultExprValueCodec{}
			err := valCodec.Decode(raw, &res)
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
	r.dao.RangeResource(ctx, rev, fmtExprPrefix, func(kvs []*mvccpb.KeyValue) {
		var exprs []*Expression
		valCodec := &defaultExprValueCodec{}
		for index := range kvs {
			var expr Expression
			err := valCodec.Decode(kvs[index].Value, &expr)
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
	r.dao.WatchResource(ctx, rev, fmtExprPrefix, func(et dao.EnventType, kv *mvccpb.KeyValue) {
		var expr Expression
		valCodec := &defaultExprValueCodec{}
		err := valCodec.Decode(kv.Value, &expr)
		if nil != err {
			log.L().Error("")
		}
		handler(et, expr)
	})
}

type RangeExpressionFunc func([]*Expression)
type WatchExpressionFunc func(dao.EnventType, Expression)
