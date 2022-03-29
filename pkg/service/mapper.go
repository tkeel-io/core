package service

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

func checkMapper(m *dao.Mapper) error {
	if m.ID == "" {
		m.ID = util.UUID("mapper")
	}

	if m.TQL == "" {
		return xerrors.ErrInvalidRequest
	}

	// check tql parse.
	tdtlIns, err := tdtl.NewTDTL(m.TQL, nil)
	if nil != err {
		log.L().Error("check mapper", zap.Error(err), zfield.TQL(m.TQL))
		return errors.Wrap(err, "parse TQL")
	}

	propKeys := make(map[string]string)
	for key := range tdtlIns.Fields() {
		propKeys[" "+key] = " " + strings.Join([]string{FieldProps, key}, sep)
	}

	for _, keys := range tdtlIns.Entities() {
		for _, key := range keys {
			segs := strings.SplitN(key, sep, 2)
			segs = append(segs[:1], append([]string{FieldProps}, segs[1:]...)...)
			propKeys[key] = strings.Join(segs, sep)
		}
	}

	// sort.
	keys := sort.StringSlice{}
	for key := range propKeys {
		keys = append(keys, key)
	}

	sort.Sort(keys)
	for index := range keys {
		key := keys[keys.Len()-index-1]
		m.TQL = strings.ReplaceAll(m.TQL, key, propKeys[key])
	}

	// check tql parse.
	_, err = tdtl.NewTDTL(m.TQL, nil)
	if nil != err {
		log.L().Error("check mapper", zap.Error(err), zfield.TQL(m.TQL))
		return errors.Wrap(xerrors.ErrInternal, "parse TQL")
	}

	return nil
}

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

	mp := dao.Mapper{
		ID:          req.Mapper.Id,
		TQL:         req.Mapper.Tql,
		Name:        req.Mapper.Name,
		Owner:       entity.Owner,
		EntityID:    req.EntityId,
		Description: req.Mapper.Description,
	}

	// TODO: 兼容v0.3, 后面去掉.
	if mp.ID == "" && req.Mapper.Name != "" {
		mp.ID = req.Mapper.Name
	}

	// check mapper.
	if err = checkMapper(&mp); nil != err {
		log.L().Error("append mapper", zfield.Eid(req.EntityId), zap.Error(err))
		return
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

func (s *EntityService) RemoveMapper(ctx context.Context, req *pb.RemoveMapperRequest) (out *pb.RemoveMapperResponse, err error) {
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

	mp := dao.Mapper{
		ID:       req.Id,
		Owner:    entity.Owner,
		EntityID: req.EntityId,
	}

	if err = s.apiManager.RemoveMapper(ctx, &mp); nil != err {
		log.L().Error("remove mapper", zfield.Eid(req.EntityId), zap.Error(err))
		return
	}

	return &pb.RemoveMapperResponse{
		Id:       mp.ID,
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: mp.EntityID,
	}, nil
}

func (s *EntityService) GetMapper(ctx context.Context, in *pb.GetMapperRequest) (out *pb.GetMapperResponse, err error) {
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

	mp := &dao.Mapper{
		ID:       in.Id,
		Owner:    entity.Owner,
		EntityID: in.EntityId,
	}

	if mp, err = s.apiManager.GetMapper(ctx, mp); nil != err {
		log.L().Error("get mapper", zfield.Eid(in.EntityId), zap.Error(err))
		return
	}

	return &pb.GetMapperResponse{
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

func (s *EntityService) ListMapper(ctx context.Context, in *pb.ListMapperRequest) (out *pb.ListMapperResponse, err error) {
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

	var mps []dao.Mapper
	if mps, err = s.apiManager.ListMapper(ctx, &entity); nil != err {
		log.L().Error("list mapper", zfield.Eid(in.EntityId), zap.Error(err))
		return
	}

	var mpDtos = make([]*pb.Mapper, len(mps))
	for index := range mps {
		mpDtos[index] = &pb.Mapper{
			Id:          mps[index].ID,
			Name:        mps[index].Name,
			Tql:         mps[index].TQL,
			Description: mps[index].Description,
		}
	}

	return &pb.ListMapperResponse{
		Type:     entity.Type,
		Owner:    entity.Owner,
		Source:   entity.Source,
		EntityId: in.EntityId,
		Mappers:  mpDtos,
	}, nil
}
