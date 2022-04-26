package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func (s *EntityService) AppendMapper(ctx context.Context, req *pb.AppendMapperRequest) (out *pb.AppendMapperResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = req.EntityId
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, &entity)

	mp := mapper.Mapper{
		ID:          req.Mapper.Id,
		TQL:         req.Mapper.Tql,
		Name:        req.Mapper.Name,
		Owner:       entity.Owner,
		EntityID:    req.EntityId,
		Description: req.Mapper.Description,
	}

	// append mapper.
	if err = s.apiManager.AppendMapper(ctx, &mp); nil != err {
		log.L().Error("append mapper", zfield.Eid(req.EntityId), zap.Error(err))
		return
	}

	return &pb.AppendMapperResponse{
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: mp.EntityID,
		Mapper: &pb.Mapper{
			Id:          mp.ID,
			Name:        mp.Name,
			Tql:         mp.TQL,
			Description: mp.Description,
		},
	}, nil
}
