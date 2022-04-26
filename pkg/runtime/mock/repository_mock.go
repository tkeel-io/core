package mock

import (
	"context"

	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func NewRepo() repository.IRepository {
	daoIns, _ := dao.NewMock(context.Background(), config.Metadata{}, config.EtcdConfig{})
	return repository.New(daoIns)
}
