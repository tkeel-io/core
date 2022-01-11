package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/service/mock"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"google.golang.org/protobuf/types/known/structpb"
)

var entityManager entities.EntityManager
var entityService *EntityService

func TestMain(m *testing.M) {
	var err error

	// logger initialized.
	log.InitLogger("core-service", "DEBUG", true)

	entityManager = mock.NewEntityManagerMock()
	entityService, err = NewEntityService(context.Background(), entityManager, mock.NewSearchMock())
	if nil != err {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func Test_entity2EntityResponse(t *testing.T) {
	base := statem.Base{
		ID:           "device123",
		Type:         "DEVICE",
		Owner:        "admin",
		Source:       "dm",
		Version:      0,
		LastTime:     time.Now().UnixMilli(),
		Mappers:      []statem.MapperDesc{{Name: "mapper123", TQLString: "insert into device123 select device234.temp as temp"}},
		KValues:      map[string]constraint.Node{"temp": constraint.NewNode(25)},
		ConfigsBytes: nil,
	}

	out := entityService.entity2EntityResponse(&base)
	assert.Equal(t, base.ID, out.Id)
	assert.Equal(t, base.Type, out.Type)
	assert.Equal(t, base.Owner, out.Owner)
	assert.Equal(t, base.Source, out.Source)
}

func Test_CreateEntity(t *testing.T) {
	m := map[string]interface{}{}
	properties, err := structpb.NewValue(m)
	assert.Nil(t, err, "properties NewValue")
	_, err = entityService.CreateEntity(context.Background(), &pb.CreateEntityRequest{
		Id:         "device123",
		From:       "",
		Source:     "dm",
		Owner:      "admin",
		Type:       "DEVICE",
		Properties: properties,
	})
	assert.Nil(t, err)
}

func Test_UpdateEntity(t *testing.T) {
	m := map[string]interface{}{}
	properties, err := structpb.NewValue(m)
	assert.Nil(t, err, "properties NewValue")
	_, err = entityService.UpdateEntity(context.Background(), &pb.UpdateEntityRequest{
		Id:         "device123",
		Source:     "dm",
		Owner:      "admin",
		Type:       "DEVICE",
		Properties: properties,
	})
	assert.Nil(t, err)
}

func Test_PatchEntity(t *testing.T) {
	m := []interface{}{}
	properties, err := structpb.NewValue(m)
	assert.Nil(t, err, "properties NewValue")
	_, err = entityService.PatchEntity(context.Background(), &pb.PatchEntityRequest{
		Id:         "device123",
		Owner:      "admin",
		Type:       "DEVICE",
		Source:     "dm",
		Properties: properties,
	})
	assert.Nil(t, err)
}

func Test_DeleteEntity(t *testing.T) {
	_, err := entityService.DeleteEntity(context.Background(), &pb.DeleteEntityRequest{
		Id:     "device123",
		Owner:  "admin",
		Type:   "DEVICE",
		Source: "dm",
	})
	assert.Nil(t, err)
}

func Test_GetEntityProps(t *testing.T) {
	_, err := entityService.GetEntityProps(context.Background(), &pb.GetEntityPropsRequest{
		Id:     "device123",
		Owner:  "admin",
		Type:   "DEVICE",
		Source: "dm",
		Pids:   "temp,metrics.cpu",
	})
	assert.Nil(t, err)
}

func Test_GetEntity(t *testing.T) {
	_, err := entityService.GetEntity(context.Background(), &pb.GetEntityRequest{
		Id:     "device123",
		Owner:  "admin",
		Type:   "DEVICE",
		Source: "dm",
	})
	assert.Nil(t, err)
}

func Test_ListEntity(t *testing.T) {
	_, err := entityService.ListEntity(context.Background(), &pb.ListEntityRequest{
		Owner:  "admin",
		Source: "dm",
	})
	assert.Nil(t, err)
}

func Test_AppendMapper(t *testing.T) {
	_, err := entityService.AppendMapper(context.Background(), &pb.AppendMapperRequest{
		Id:     "device123",
		Owner:  "admin",
		Type:   "DEVICE",
		Source: "dm",
		Mapper: &pb.MapperDesc{
			Name: "mapper123",
			Tql:  "insert into device123 select device234.temp as temp",
		},
	})
	assert.Nil(t, err)
}

func Test_RemoveMapper(t *testing.T) {
	_, err := entityService.RemoveMapper(context.Background(), &pb.RemoveMapperRequest{
		Id:         "device123",
		Owner:      "admin",
		Type:       "DEVICE",
		Source:     "dm",
		MapperName: "mapper123",
	})
	assert.Nil(t, err)
}
