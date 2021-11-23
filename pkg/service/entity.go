package service

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/statem"

	"google.golang.org/protobuf/types/known/structpb"
)

type EntityService struct {
	pb.UnimplementedEntityServer
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
	searchClient  pb.SearchHTTPServer
}

func NewEntityService(ctx context.Context, mgr *entities.EntityManager, searchClient pb.SearchHTTPServer) (*EntityService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &EntityService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
		searchClient:  searchClient,
	}, nil
}

func (s *EntityService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	if req.Id != "" {
		entity.ID = req.Id
	}
	entity.Owner = req.Owner
	entity.Type = req.Type
	entity.Source = req.Plugin
	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
	case nil:
		log.Warn("empty params")
	default:
		log.Errorf("create entity failed, invalid params, %v", kv)
		return out, ErrEntityInvalidParams
	}

	// set properties.
	_, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2EntityResponse(entity)
	return
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Plugin
	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
	case nil:
		log.Warn("empty params")
	default:
		return
	}

	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2EntityResponse(entity)

	return
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (out *pb.DeleteEntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Plugin

	// delete entity.
	_, err = s.entityManager.DeleteEntity(ctx, entity)
	if nil != err {
		return
	}

	out = &pb.DeleteEntityResponse{}
	out.Id = req.Id
	out.Status = "ok"
	return
}

func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Plugin

	// get entity from entity manager.
	entity, err = s.entityManager.GetProperties(ctx, entity)
	if nil != err {
		log.Errorf("get entity failed, %s", err.Error())
		return
	}

	out = s.entity2EntityResponse(entity)

	return
}

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (out *pb.ListEntityResponse, err error) {
	searchReq := &pb.SearchRequest{}
	searchReq.Query = req.Query
	searchReq.Page = req.Page
	searchReq.Condition = req.Condition

	resp, err := s.searchClient.Search(ctx, searchReq)
	if err != nil {
		log.Errorf("list entities failed, err: %s", err.Error())
		return
	}

	out = &pb.ListEntityResponse{}
	out.Total = resp.Total
	out.Limit = resp.Limit
	out.Items = make([]*pb.EntityResponse, resp.Limit)
	for _, item := range resp.Items {
		switch kv := item.AsInterface().(type) {
		case map[string]interface{}:
			properties, _ := structpb.NewValue(kv)
			entityItem := &pb.EntityResponse{
				Id:         interface2string(kv["id"]),
				Plugin:     "",
				Source:     "",
				Owner:      "",
				Type:       "",
				Properties: properties,
				Mappers:    []*pb.MapperDesc{},
			}
			out.Items = append(out.Items, entityItem)
		}
	}
	if err != nil {
		log.Errorf("list entities failed, err: %s", err.Error())
		return out, errors.Wrap(err, "entity search failed")
	}
	return out, nil
}

func (s *EntityService) entity2EntityResponse(entity *Entity) (out *pb.EntityResponse) {
	if entity == nil {
		return
	}

	var err error
	out = &pb.EntityResponse{}

	kv := make(map[string]interface{})
	for k, v := range entity.KValues {
		kv[k] = v.Value()
	}

	configs := make(map[string]interface{})
	bytes, _ := json.Marshal(entity.Configs)
	json.Unmarshal(bytes, &configs)

	if out.Properties, err = structpb.NewValue(kv); nil != err {
		log.Errorf("convert entity failed, err: %s", err)
	} else if out.Configs, err = structpb.NewValue(configs); nil != err {
		log.Errorf("convert entity failed, err: %s", err)
	}

	out.Mappers = make([]*pb.MapperDesc, 0)

	for _, mapper := range entity.Mappers {
		out.Mappers = append(out.Mappers, &pb.MapperDesc{Name: mapper.Name, Tql: mapper.TQLString})
	}

	out.Plugin = entity.Source
	out.Owner = entity.Owner
	out.Id = entity.ID
	out.Type = entity.Type

	return out
}

func (s *EntityService) AppendMapper(ctx context.Context, req *pb.AppendMapperRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Plugin

	mapperDesc := statem.MapperDesc{}
	if req.Mapper != nil {
		mapperDesc.Name = req.Mapper.Name
		mapperDesc.TQLString = req.Mapper.Tql
		entity.Mappers = []statem.MapperDesc{mapperDesc}
	} else {
		return nil, errors.Wrap(ErrEntityMapperNil, "append mapper to entity failed")
	}
	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2EntityResponse(entity)
	return
}

func (s *EntityService) SetEntityConfigs(ctx context.Context, req *pb.SetEntityConfigRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Plugin

	entity.Configs, err = parseConfigFrom(ctx, req.Configs.AsInterface())
	if nil != err {
		log.Errorf("set entity config failed, %s", err.Error())
		return out, err
	}

	// set properties.
	entity, err = s.entityManager.SetConfigs(ctx, entity)
	if nil != err {
		log.Errorf("set entity config failed, %s", err.Error())
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "entity set config failed")
}

func parseConfigFrom(ctx context.Context, data interface{}) (out map[string]constraint.Config, err error) {
	// parse configs from.
	out = make(map[string]constraint.Config)
	switch configs := data.(type) {
	case []interface{}:
		for _, cfg := range configs {
			if c, ok := cfg.(map[string]interface{}); ok {
				var cfgRet constraint.Config
				if cfgRet, err = constraint.ParseConfigsFrom(c); nil != err {
					return out, errors.Wrap(err, "parse entity config failed")
				}
				out[cfgRet.ID] = cfgRet
				continue
			}
			return out, ErrEntityConfigInvalid
		}
	default:
		return out, ErrEntityConfigInvalid
	}
	return out, errors.Wrap(err, "parse entity config failed")
}
