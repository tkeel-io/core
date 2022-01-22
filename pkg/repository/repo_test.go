package repository

// import (
// 	"context"
// 	"testing"

// 	"github.com/tkeel-io/core/pkg/config"
// 	"github.com/tkeel-io/core/pkg/repository/dao"
// )

// var coreDao *dao.Dao

// func TestMain(m *testing.M) {
// 	var err error
// 	coreDao, err = dao.New(
// 		context.Background(),
// 		config.Metadata{Name: "dapr", Properties: []config.Pair{{Key: "store_name", Value: "core-entity"}}},
// 		config.EtcdConfig{Endpoints: []string{"http://localhost:2379"}, DialTimeout: 3},
// 	)

// 	if nil != err {
// 		panic(err)
// 	}
// }

// func TestDao_GetEntity(t *testing.T) {

// }
