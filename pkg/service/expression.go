package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
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

	// append expressions.
	expressions := make([]dao.Expression, len(req.Expressions.Expressions))
	for index, expr := range req.Expressions.Expressions {
		expressions[index] = *dao.NewExpression(
			req.Owner, req.EntityId, propKey(expr.Path), expr.Expression)
	}

	if err = s.apiManager.AppendExpression(ctx, expressions); nil != err {
		log.L().Error("append expressions",
			zfield.Eid(req.EntityId), zfield.Owner(req.Owner), zap.Error(err))
	}

	return &pb.AppendExpressionResp{
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: entity.ID,
		Count:    int32(len(expressions)),
	}, nil
}

func (s *EntityService) RemoveExpression(ctx context.Context, req *pb.RemoveExpressionReq) (out *pb.RemoveExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	log.L().Debug("remove expression", zfield.Owner(req.Owner),
		zfield.Eid(req.EntityId), zfield.Path(req.Path))

	var entity Entity
	entity.ID = req.EntityId
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, &entity)

	if err = s.apiManager.RemoveExpression(ctx, dao.Expression{
		Path:     propKey(req.Path),
		Owner:    req.Owner,
		EntityID: req.EntityId,
	}); nil != err {
		log.L().Error("remove expression", zfield.Owner(req.Owner),
			zfield.Eid(req.EntityId), zfield.Path(req.Path))
		return nil, errors.Wrap(err, "remove expression")
	}

	return &pb.RemoveExpressionResp{
		Path:     req.Path,
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: req.EntityId,
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

	var expr *dao.Expression
	if expr, err = s.apiManager.GetExpression(ctx,
		dao.Expression{
			Path:     in.Path,
			Owner:    entity.Owner,
			EntityID: in.EntityId,
		}); nil != err {
		log.L().Error("get expression", zap.Error(err),
			zfield.Eid(in.EntityId), zfield.Owner(entity.Owner), zfield.Path(in.Path))
		return nil, errors.Wrap(err, "get expression")
	}

	return &pb.GetExpressionResp{
		Type:       entity.Type,
		Owner:      entity.Owner,
		Source:     entity.Source,
		EntityId:   expr.EntityID,
		Expression: dao2pbExpression(expr),
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

	var exprs []dao.Expression
	if exprs, err = s.apiManager.ListExpression(ctx,
		&apim.Base{
			ID:    in.EntityId,
			Owner: in.Owner,
		}); nil != err {
		log.L().Error("list expressions", zap.Error(err),
			zfield.Eid(in.EntityId), zfield.Owner(in.Owner))
		return nil, errors.Wrap(err, "list expressions")
	}

	out = &pb.ListExpressionResp{
		Type:        entity.Type,
		Owner:       entity.Owner,
		Source:      entity.Source,
		EntityId:    in.EntityId,
		Expressions: []*pb.Expression{},
	}

	for index := range exprs {
		out.Expressions = append(out.Expressions,
			dao2pbExpression(&exprs[index]))
	}

	return out, nil
}

func dao2pbExpression(expr *dao.Expression) *pb.Expression {
	path := ""
	if expr.Type == dao.ExprTypeEval {
		path = expr.Path
	}

	return &pb.Expression{
		Path:        path,
		Name:        expr.Name,
		Expression:  expr.Expression,
		Description: expr.Description,
	}
}
