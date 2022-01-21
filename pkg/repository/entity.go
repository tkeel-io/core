package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func (r *repo) PutEntity(ctx context.Context, en *dao.Entity) error {
	return errors.Wrap(r.dao.PutEntity(ctx, en), "put entity repository")
}

func (r *repo) GetEntity(ctx context.Context, en *dao.Entity) (*dao.Entity, error) {
	en, err := r.dao.GetEntity(ctx, en.ID)
	return en, errors.Wrap(err, "put entity repository")
}

func (r *repo) DelEntity(ctx context.Context, en *dao.Entity) error {
	return errors.Wrap(r.dao.DelEntity(ctx, en.ID), "del entity repository")
}

func (r *repo) HasEntity(ctx context.Context, en *dao.Entity) (bool, error) {
	has, err := r.dao.HasEntity(ctx, en.ID)
	return has, errors.Wrap(err, "exists entity repository")
}
