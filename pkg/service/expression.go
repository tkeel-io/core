package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

func (s *EntityService) AppendExpression(ctx context.Context, req *pb.AppendExpressionReq) (out *pb.AppendExpressionResp, err error) {
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

	expr := dao.Expression{
		ID:          util.UUID("expr"),
		Name:        req.Expression.Name,
		Owner:       entity.Owner,
		EntityID:    req.EntityId,
		Expression:  req.Expression.Expression,
		Description: req.Expression.Description,
	}

	// append expression.

	return &pb.AppendExpressionResp{
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: expr.EntityID,
		Expression: &pb.Expression{
			Id:          expr.ID,
			Name:        expr.Name,
			Expression:  expr.Expression,
			Description: expr.Description,
		},
	}, nil
}

func (s *EntityService) RemoveExpression(ctx context.Context, req *pb.RemoveExpressionReq) (out *pb.RemoveExpressionResp, err error) {
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

	expr := dao.Expression{
		ID:       req.Id,
		Owner:    entity.Owner,
		EntityID: req.EntityId,
	}

	return &pb.RemoveExpressionResp{
		Id:       expr.ID,
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: expr.EntityID,
	}, nil
}

func (s *EntityService) GetExpression(ctx context.Context, in *pb.GetExpressionReq) (out *pb.GetExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = in.EntityId
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, &entity)

	expr := &dao.Expression{
		ID:       in.Id,
		Owner:    entity.Owner,
		EntityID: in.EntityId,
	}

	return &pb.GetExpressionResp{
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: expr.EntityID,
		Expression: &pb.Expression{
			Id:          expr.ID,
			Name:        expr.Name,
			Expression:  expr.Expression,
			Description: expr.Description,
		},
	}, nil
}

func (s *EntityService) ListExpression(ctx context.Context, in *pb.ListExpressionReq) (out *pb.ListExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity Entity
	entity.ID = in.EntityId
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, &entity)

	return &pb.ListExpressionResp{
		Type:        entity.Type,
		Owner:       entity.Owner,
		Source:      entity.Source,
		EntityId:    in.EntityId,
		Expressions: []*pb.Expression{},
	}, nil
}
