package service

import (
	"context"
	"errors"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"

	"google.golang.org/protobuf/types/known/structpb"
)

type EntityService struct {
	pb.UnimplementedEntityServer
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

func NewEntityService(ctx context.Context, mgr *entities.EntityManager) (*EntityService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &EntityService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
}

func (s *EntityService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	if req.Id != "" {
		entity.ID = req.Id
	}
	entity.Owner = req.Owner
	entity.Type = req.Type
	entity.PluginID = req.Plugin
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		entity.KValues = kv
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

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		entity.KValues = kv
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
	entity.PluginID = req.Plugin

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
	entity.PluginID = req.Plugin

	// get entity from entity manager.
	entity, err = s.entityManager.GetAllProperties(ctx, entity)
	if nil != err {
		log.Errorf("get entity failed, %s", err.Error())
		return
	}

	out = s.entity2EntityResponse(entity)

	return
}

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (*pb.ListEntityResponse, error) {
	return &pb.ListEntityResponse{}, nil
}

func (s *EntityService) entity2EntityResponse(entity *Entity) (out *pb.EntityResponse) {
	if entity == nil {
		return
	}

	out = &pb.EntityResponse{}
	out.Properties, _ = structpb.NewValue(entity.KValues)
	out.Mappers = make([]*pb.MapperDesc, 0)

	for _, mapper := range entity.Mappers {
		out.Mappers = append(out.Mappers, &pb.MapperDesc{Name: mapper.Name, Tql: mapper.TQLString})
	}

	out.Plugin = entity.PluginID
	out.Owner = entity.Owner
	out.Id = entity.ID
	out.Type = entity.Type
	return out
}

func (s *EntityService) AppendMapper(ctx context.Context, req *pb.AppendMapperRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin

	mapperDesc := entities.MapperDesc{}
	if req.Mapper != nil {
		mapperDesc.Name = req.Mapper.Name
		mapperDesc.TQLString = req.Mapper.Tql
		entity.Mappers = []entities.MapperDesc{mapperDesc}
	} else {
		return nil, errors.New("mapper is nil")
	}
	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2EntityResponse(entity)
	return
}
