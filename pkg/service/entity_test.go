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
	"google.golang.org/protobuf/types/known/structpb"
)

var entityManager entities.EntityManager
var entityService *EntityService

func TestMain(m *testing.M) {
	var err error
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
