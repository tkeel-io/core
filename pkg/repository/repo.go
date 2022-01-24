package repository

import (
	"context"

	"github.com/tkeel-io/core/pkg/repository/dao"
)

type repo struct {
	dao *dao.Dao
}

func New(dao *dao.Dao) IRepository {
	return &repo{dao: dao}
}

func (r *repo) GetLastRevision(ctx context.Context) int64 {
	return r.dao.GetLastRevision(ctx)
}
