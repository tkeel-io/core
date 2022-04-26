package repository

import (
	"context"

	"github.com/tkeel-io/core/pkg/repository/dao"
)

type repo struct {
	dao dao.IDao
}

func New(dao dao.IDao) IRepository {
	return &repo{dao: dao}
}

func (r *repo) GetLastRevision(ctx context.Context) int64 {
	return r.dao.GetLastRevision(ctx)
}
