package repository

import (
	"github.com/tkeel-io/core/pkg/repository/dao"
)

type repo struct {
	dao *dao.Dao
}

func New(dao *dao.Dao) IRepository {
	return &repo{dao: dao}
}
