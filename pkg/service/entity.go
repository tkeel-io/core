package service

import (
	"context"
	"errors"

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
	entity.Source = req.Plugin
	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
	default:
		return
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
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
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

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (*pb.ListEntityResponse, error) {
	return &pb.ListEntityResponse{}, nil
}

func (s *EntityService) entity2EntityResponse(entity *Entity) (out *pb.EntityResponse) {
	if entity == nil {
		return
	}

	out = &pb.EntityResponse{}

	kv := make(map[string]interface{})
	for k, v := range entity.KValues {
		kv[k] = v.Value()
	}

	out.Properties, _ = structpb.NewValue(kv)
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
