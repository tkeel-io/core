package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

const (
	SchemaPrefix = "/core/v1/schema"
)

type ListSchemaReq struct {
	Owner    string
	EntityID string
}

var _ dao.Resource = (*Schema)(nil)

type Schema struct {
	// Schema identifier.
	ID string
	// Schema name.
	Name string
	// Schema owner.
	Owner string
	// Schema.
	Schema string
	// description.
	Description string
}

func NewSchema(owner, ID, name, schema, desc string) *Schema {
	return &Schema{
		ID:          ID,
		Name:        name,
		Owner:       owner,
		Schema:      schema,
		Description: desc,
	}
}

func ListSchemaPrefix(Owner, EntityID string) string {
	keyString := fmt.Sprintf("%s/%s",
		SchemaPrefix, Owner)
	return keyString
}

func (s *Schema) EncodeKey() ([]byte, error) {
	if s.Owner == ""{
		return nil, errors.Errorf("Schema Owner is empty")
	}
	if s.ID == ""{
		return nil, errors.Errorf("Schema ID is empty")
	}

	keyString := fmt.Sprintf("%s/%s/%s",
		SchemaPrefix, s.Owner, s.ID)
	return []byte(keyString), nil
}

func (s *Schema) Encode() ([]byte, error) {
	bytes, err := json.Marshal(s)
	return bytes, errors.Wrap(err, "encode Schema")
}

func (s *Schema) Decode(bytes []byte) error {
	err := json.Unmarshal(bytes, s)
	return errors.Wrap(err, "decode Schema")
}

func (r *repo) PutSchema(ctx context.Context, expr Schema) error {
	err := r.dao.PutResource(ctx, &expr)
	return errors.Wrap(err, "put expression repository")
}

func (r *repo) GetSchema(ctx context.Context, expr Schema) (Schema, error) {
	_, err := r.dao.GetResource(ctx, &expr)
	return expr, errors.Wrap(err, "get expression repository")
}

func (r *repo) DelSchema(ctx context.Context, expr Schema) error {
	err := r.dao.DelResource(ctx, &expr)
	return errors.Wrap(err, "del expression repository")
}

func (r *repo) HasSchema(ctx context.Context, expr Schema) (bool, error) {
	has, err := r.dao.HasResource(ctx, &expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListSchema(ctx context.Context, rev int64, req *ListSchemaReq) ([]*Schema, error) {
	// construct prefix.
	prefix := ListSchemaPrefix(req.EntityID, req.Owner)
	ress, err := r.dao.ListResource(ctx, rev, prefix,
		func(raw []byte) (dao.Resource, error) {
			var res Schema // escape.
			err := res.Decode(raw)
			return &res, errors.Wrap(err, "decode expression")
		})

	var exprs []*Schema
	for index := range ress {
		if expr, ok := ress[index].(*Schema); ok {
			exprs = append(exprs, expr)
			continue
		}
		// panic.
	}
	return exprs, errors.Wrap(err, "list expression repository")
}

func (r *repo) RangeSchema(ctx context.Context, rev int64, handler RangeSchemaFunc) {
	r.dao.RangeResource(ctx, rev, SchemaPrefix, func(kvs []*mvccpb.KeyValue) {
		var exprs []*Schema
		for index := range kvs {
			var expr Schema
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

func (r *repo) WatchSchema(ctx context.Context, rev int64, handler WatchSchemaFunc) {
	r.dao.WatchResource(ctx, rev, SchemaPrefix, func(et dao.EnventType, kv *mvccpb.KeyValue) {
		var expr Schema
		err := expr.Decode(kv.Value)
		if nil != err {
			log.L().Error("")
		}
		handler(et, expr)
	})
}

type RangeSchemaFunc func([]*Schema)
type WatchSchemaFunc func(dao.EnventType, Schema)
