package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/service/mock"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"google.golang.org/protobuf/types/known/structpb"
)

var apiManager apim.APIManager
var entityService *EntityService

func TestMain(m *testing.M) {
	var err error

	// logger initialized.
	log.InitLogger("core-service", "DEBUG", true)

	searchMock := mock.NewSearchMock()
	apiManager = mock.NewAPIManagerMock()
	entityService, err = NewEntityService(context.Background())
	if nil != err {
		os.Exit(1)
	}

	entityService.Init(apiManager, searchMock)
	os.Exit(m.Run())
}

func Test_entity2EntityResponse(t *testing.T) {
	base := apim.Base{
		ID:         "device123",
		Type:       "DEVICE",
		Owner:      "admin",
		Source:     "dm",
		Version:    0,
		LastTime:   util.UnixMilli(),
		Mappers:    []state.Mapper{{Name: "mapper123", TQL: "insert into device123 select device234.temp as temp"}},
		Properties: map[string]tdtl.Node{"temp": tdtl.IntNode(25)},
		ConfigFile: nil,
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
	// 	m := map[string]interface{}{}
	// 	properties, err := structpb.NewValue(m)
	// 	assert.Nil(t, err, "properties NewValue")
	// 	_, err = entityService.UpdateEntity(context.Background(), &pb.UpdateEntityRequest{
	// 		Id:         "device123",
	// 		Source:     "dm",
	// 		Owner:      "admin",
	// 		Type:       "DEVICE",
	// 		Properties: properties,
	// 	})
	// 	assert.NotNil(t, err)
}

func Test_PatchEntity(t *testing.T) {
	m := []interface{}{}
	properties, err := structpb.NewValue(m)
	assert.Nil(t, err, "properties NewValue")
	_, err = entityService.PatchEntityProps(context.Background(), &pb.PatchEntityPropsRequest{
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
		Id:           "device123",
		Owner:        "admin",
		Type:         "DEVICE",
		Source:       "dm",
		PropertyKeys: "temp,metrics.cpu",
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
		EntityId: "device123",
		Owner:    "admin",
		Type:     "DEVICE",
		Source:   "dm",
		Mapper: &pb.Mapper{
			Id:          "mapper123",
			Name:        "mapper123",
			Tql:         "insert into device123 select device234.temp as temp",
			Description: "test mapper.",
		},
	})
	assert.Nil(t, err)
}

func Test_RemoveMapper(t *testing.T) {
	_, err := entityService.RemoveMapper(context.Background(), &pb.RemoveMapperRequest{
		Id:       "mapper123",
		Owner:    "admin",
		Type:     "DEVICE",
		Source:   "dm",
		EntityId: "device123",
	})
	assert.Nil(t, err)
}

func Test_SetConfigs(t *testing.T) {
	configs := map[string]interface{}{
		"configs1": []interface{}{
			map[string]interface{}{
				"id":   "temp",
				"type": "int",
				"define": map[string]interface{}{
					"max":  100,
					"unit": "°",
					"ext": map[string]interface{}{
						"unit_zh": "度",
					},
				},
			},
		},
		"configs2": []interface{}{
			map[string]interface{}{
				"id":   "temps",
				"type": "array",
				"define": map[string]interface{}{
					"length": 20,
					"elem_type": map[string]interface{}{
						"id":   "temp",
						"type": "int",
						"define": map[string]interface{}{
							"max":  100,
							"unit": "°",
							"ext": map[string]interface{}{
								"unit_zh": "度",
							},
						},
					},
				},
			},
		},
		"configs3": []interface{}{
			map[string]interface{}{
				"id":   "metrics",
				"type": "struct",
				"define": map[string]interface{}{
					"fields": map[string]interface{}{
						"temp": map[string]interface{}{
							"id":   "temp",
							"type": "float",
							"define": map[string]interface{}{
								"max":  100,
								"unit": "°",
								"ext": map[string]interface{}{
									"unit_zh": "度",
								},
							},
						},
						"cpu_used": map[string]interface{}{
							"id":          "cpu_used",
							"type":        "float",
							"enabled":     true,
							"description": "cpu使用率",
							"define": map[string]interface{}{
								"max": 1,
								"min": 0,
							},
						},
						"mem_used": map[string]interface{}{
							"id":          "mem_used",
							"type":        "float",
							"enabled":     true,
							"description": "内存使用率",
							"define": map[string]interface{}{
								"max": 1,
								"min": 0,
							},
						},
					},
				},
			},
		},
	}

	for name, cfg := range configs {
		t.Run(name, func(t *testing.T) {
			c, err := structpb.NewValue(cfg)
			assert.Nil(t, err)

			res, err := entityService.UpdateEntityConfigs(context.Background(), &pb.UpdateEntityConfigsRequest{
				Id:      "device123",
				Owner:   "admin",
				Type:    "DEVICE",
				Source:  "dm",
				Configs: c,
			})
			assert.Nil(t, err)
			assert.Equal(t, "device123", res.Id)
			assert.Equal(t, "admin", res.Owner)
			assert.Equal(t, "DEVICE", res.Type)
		})
	}
}

func Test_RemoveEntityConfigs(t *testing.T) {
	res, err := entityService.RemoveEntityConfigs(context.Background(),
		&pb.RemoveEntityConfigsRequest{
			Id:           "device123",
			Owner:        "admin",
			Type:         "DEVICE",
			Source:       "dm",
			PropertyKeys: "temp,metrics.cpu_used",
		})
	assert.Nil(t, err)
	assert.Equal(t, "device123", res.Id)
	assert.Equal(t, "admin", res.Owner)
	assert.Equal(t, "DEVICE", res.Type)
}

func Test_GetEntityConfigs(t *testing.T) {
	res, err := entityService.GetEntityConfigs(context.Background(), &pb.GetEntityConfigsRequest{
		Id:           "device123",
		Owner:        "admin",
		Type:         "DEVICE",
		Source:       "dm",
		PropertyKeys: "temp,metrics.cpu_used",
	})
	assert.Nil(t, err)
	assert.Equal(t, "device123", res.Id)
	assert.Equal(t, "admin", res.Owner)
	assert.Equal(t, "DEVICE", res.Type)
}

func Test_PatchEntityConfigs(t *testing.T) {
	configs, err := structpb.NewValue([]interface{}{
		map[string]interface{}{
			"path":     "metrics.cpu_used",
			"operator": "relpace",
			"value": map[string]interface{}{
				"type": "int",
				"define": map[string]interface{}{
					"max":  100,
					"unit": "°",
					"ext": map[string]interface{}{
						"unit_zh": "度",
					},
				},
			},
		},
		map[string]interface{}{
			"path":     "metrics.mem_used",
			"operator": "add",
			"value": map[string]interface{}{
				"type": "int",
				"define": map[string]interface{}{
					"max":  100,
					"unit": "°",
					"ext": map[string]interface{}{
						"unit_zh": "度",
					},
				},
			},
		},
		map[string]interface{}{
			"path":     "metrics.temp",
			"operator": "remove",
			"value": map[string]interface{}{
				"type": "int",
				"define": map[string]interface{}{
					"max":  100,
					"unit": "°",
					"ext": map[string]interface{}{
						"unit_zh": "度",
					},
				},
			},
		},
		map[string]interface{}{
			"path":     "metrics.cpu_num",
			"operator": "copy",
			"value": map[string]interface{}{
				"type": "int",
				"define": map[string]interface{}{
					"max":  100,
					"unit": "°",
					"ext": map[string]interface{}{
						"unit_zh": "度",
					},
				},
			},
		},
	})
	assert.Nil(t, err)

	res, err := entityService.PatchEntityConfigs(context.Background(),
		&pb.PatchEntityConfigsRequest{
			Id:      "device123",
			Owner:   "admin",
			Type:    "DEVICE",
			Source:  "dm",
			Configs: configs,
		})
	assert.Nil(t, err)
	assert.Equal(t, "device123", res.Id)
	assert.Equal(t, "admin", res.Owner)
	assert.Equal(t, "DEVICE", res.Type)
}
