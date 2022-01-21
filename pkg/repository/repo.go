package repository

import (
	"github.com/tkeel-io/core/pkg/repository/dao"
)

type repo struct {
	dao *dao.Dao
}

func New() IRepository {
	return &repo{}
}
