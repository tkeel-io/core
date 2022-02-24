package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func (s *EntityService) AppendMapper(ctx context.Context, req *pb.AppendMapperRequest) (out *pb.AppendMapperResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = req.EntityId
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, &entity)

	mp := dao.Mapper{
		ID:          req.Mapper.Id,
		TQL:         req.Mapper.TqlText,
		Name:        req.Mapper.Name,
		Owner:       entity.Owner,
		EntityID:    req.EntityId,
		Description: req.Mapper.Description,
	}

	// append mapper.
	if err = s.apiManager.AppendMapper(ctx, &mp); nil != err {
		log.Error("append mapper", zfield.Eid(req.EntityId), zap.Error(err))
		return
	}

	return &pb.AppendMapperResponse{}, nil
}

func (s *EntityService) RemoveMapper(ctx context.Context, req *pb.RemoveMapperRequest) (out *pb.RemoveMapperResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = req.EntityId
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, &entity)

	mp := dao.Mapper{
		ID:       req.Id,
		Owner:    entity.Owner,
		EntityID: req.EntityId,
	}

	if err = s.apiManager.RemoveMapper(ctx, &mp); nil != err {
		log.Error("remove mapper", zfield.Eid(req.EntityId), zap.Error(err))
		return
	}

	return &pb.RemoveMapperResponse{}, nil
}

func (s *EntityService) GetMapper(ctx context.Context, in *pb.GetMapperRequest) (out *pb.GetMapperResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = in.EntityId
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, &entity)

	mp := &dao.Mapper{
		ID:       in.Id,
		Owner:    entity.Owner,
		EntityID: in.EntityId,
	}

	if mp, err = s.apiManager.GetMapper(ctx, mp); nil != err {
		log.Error("get mapper", zfield.Eid(in.EntityId), zap.Error(err))
		return
	}

	return &pb.GetMapperResponse{
		EntityId: in.EntityId,
		Mapper: &pb.Mapper{
			Id:          mp.ID,
			Name:        mp.Name,
			TqlText:     mp.TQL,
			Description: mp.Description,
		},
	}, nil
}

func (s *EntityService) ListMapper(ctx context.Context, in *pb.ListMapperRequest) (out *pb.ListMapperResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = in.EntityId
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, &entity)

	var mps []dao.Mapper
	if mps, err = s.apiManager.ListMapper(ctx, &entity); nil != err {
		log.Error("list mapper", zfield.Eid(in.EntityId), zap.Error(err))
		return
	}

	var mpDtos = make([]*pb.Mapper, len(mps))
	for index := range mps {
		mpDtos[index] = &pb.Mapper{
			Id:          mps[index].ID,
			Name:        mps[index].Name,
			TqlText:     mps[index].TQL,
			Description: mps[index].Description,
		}
	}

	return &pb.ListMapperResponse{
		EntityId: in.EntityId,
		Mappers:  mpDtos,
	}, nil
}
