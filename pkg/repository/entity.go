package repository

import (
	"context"

	"github.com/pkg/errors"
)

func (r *repo) PutEntity(ctx context.Context, eid string, data []byte) error {
	return errors.Wrap(r.dao.PutEntity(ctx, eid, data), "put entity repository")
}

func (r *repo) GetEntity(ctx context.Context, eid string) ([]byte, error) {
	en, err := r.dao.GetEntity(ctx, eid)
	return en, errors.Wrap(err, "get entity repository")
}

func (r *repo) DelEntity(ctx context.Context, eid string) error {
	return errors.Wrap(r.dao.DelEntity(ctx, eid), "del entity repository")
}

func (r *repo) HasEntity(ctx context.Context, eid string) (bool, error) {
	has, err := r.dao.HasEntity(ctx, eid)
	return has, errors.Wrap(err, "exists entity repository")
}
