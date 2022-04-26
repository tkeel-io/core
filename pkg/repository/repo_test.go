package repository

import (
	"context"
	"os"
	"testing"

	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

var (
	coreDao dao.IDao
	rr      *repo
)

func TestMain(m *testing.M) {
	var err error
	coreDao, err := dao.New(
		context.Background(),
		config.Metadata{Name: "noop"},
		config.EtcdConfig{Endpoints: []string{"http://localhost:2379"}, DialTimeout: 3},
	)

	rr = &repo{
		dao: coreDao,
	}

	if nil != err {
		panic(err)
	}

	repoIns = rr
	os.Exit(m.Run())
	coreDao.Close()
}
