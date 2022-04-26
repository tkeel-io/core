package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func (s *EntityService) AppendExpression(ctx context.Context, req *pb.AppendExpressionReq) (out *pb.AppendExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	en := Entity{
		ID:     req.EntityId,
		Owner:  req.Owner,
		Source: req.Source}
	parseHeaderFrom(ctx, &en)

	log.L().Debug("append expression", zfield.Owner(req.Owner),
		zfield.Eid(req.EntityId), zfield.Value(req.Expressions))

	// append expressions.
	expressions := make([]repository.Expression, len(req.Expressions.Expressions))
	for index, expr := range req.Expressions.Expressions {
		expressions[index] = *repository.NewExpression(
			req.Owner, req.EntityId, expr.Name,
			propKey(expr.Path), expr.Expression, expr.Description)
	}

	if err = s.apiManager.AppendExpression(ctx, expressions); nil != err {
		log.L().Error("append expressions",
			zfield.Eid(req.EntityId), zfield.Owner(req.Owner), zap.Error(err))
	}

	return &pb.AppendExpressionResp{
		Owner:    en.Owner,
		EntityId: en.ID,
		Count:    int32(len(expressions)),
	}, nil
}

func (s *EntityService) RemoveExpression(ctx context.Context, req *pb.RemoveExpressionReq) (out *pb.RemoveExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	en := Entity{
		ID:     req.EntityId,
		Owner:  req.Owner,
		Source: req.Source}
	parseHeaderFrom(ctx, &en)

	log.L().Debug("remove expression", zfield.Owner(en.Owner),
		zfield.Eid(en.ID), zfield.Path(req.Paths))

	paths := []string{}
	if pathText := strings.TrimSpace(req.Paths); len(pathText) > 0 {
		paths = strings.Split(pathText, ",")
	}

	exprs := []repository.Expression{}
	for index := range paths {
		exprs = append(exprs,
			repository.Expression{
				Path:     propKey(paths[index]),
				Owner:    en.Owner,
				EntityID: en.ID,
			})
	}

	if err = s.apiManager.RemoveExpression(ctx, exprs); nil != err {
		log.L().Error("remove expressions",
			zfield.Owner(en.Owner), zfield.Eid(en.ID), zfield.Path(req.Paths))
		return nil, errors.Wrap(err, "remove expressions")
	}

	return &pb.RemoveExpressionResp{
		EntityId: en.ID,
		Owner:    en.Owner,
		Count:    int32(len(exprs)),
	}, nil
}

func (s *EntityService) GetExpression(ctx context.Context, in *pb.GetExpressionReq) (out *pb.GetExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	en := Entity{
		ID:     in.EntityId,
		Owner:  in.Owner,
		Source: in.Source}
	parseHeaderFrom(ctx, &en)

	var expr *repository.Expression
	if expr, err = s.apiManager.GetExpression(ctx,
		repository.Expression{
			Path:     propKey(in.Path),
			Owner:    en.Owner,
			EntityID: en.ID,
		}); nil != err {
		log.L().Error("get expression", zap.Error(err),
			zfield.Eid(in.EntityId), zfield.Owner(en.Owner), zfield.Path(in.Path))
		return nil, errors.Wrap(err, "get expression")
	}

	return &pb.GetExpressionResp{
		Owner:      en.Owner,
		EntityId:   expr.EntityID,
		Expression: dao2pbExpression(expr),
	}, nil
}

func (s *EntityService) ListExpression(ctx context.Context, in *pb.ListExpressionReq) (out *pb.ListExpressionResp, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.EntityId))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	en := Entity{
		ID:     in.EntityId,
		Owner:  in.Owner,
		Source: in.Source}
	parseHeaderFrom(ctx, &en)

	var exprs []*repository.Expression
	if exprs, err = s.apiManager.ListExpression(ctx,
		&apim.Base{
			ID:    en.ID,
			Owner: en.Owner,
		}); nil != err {
		log.L().Error("list expressions", zap.Error(err),
			zfield.Eid(en.ID), zfield.Owner(en.Owner))
		return nil, errors.Wrap(err, "list expressions")
	}

	out = &pb.ListExpressionResp{
		Owner:       en.Owner,
		EntityId:    en.ID,
		Expressions: []*pb.Expression{},
	}

	for index := range exprs {
		out.Expressions = append(out.Expressions,
			dao2pbExpression(exprs[index]))
	}

	return out, nil
}

func dao2pbExpression(expr *repository.Expression) *pb.Expression {
	var path string
	if expr.Type == repository.ExprTypeEval {
		path = expr.Path
		segs := strings.SplitN(expr.Path, sep, 2)
		if len(segs) == 2 {
			path = segs[1]
		}
	}

	return &pb.Expression{
		Path:        path,
		Name:        expr.Name,
		Expression:  expr.Expression,
		Description: expr.Description,
	}
}
