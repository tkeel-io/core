/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"

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
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
	case nil:
		log.Warn("create entity, but empty params", logger.EntityID(req.Id))
	default:
		log.Error("create entity, but invalid params",
			logger.EntityID(req.Id), zap.Error(ErrEntityInvalidParams))
		return out, ErrEntityInvalidParams
	}

	// check properties.
	if _, has := entity.KValues[""]; has {
		log.Error("create entity, but invalid params",
			logger.EntityID(req.Id), zap.Error(ErrEntityPropertyIDEmpty))
		return out, ErrEntityPropertyIDEmpty
	}

	// set template entity id.
	ctx = context.WithValue(ctx, entities.TemplateEntityID{}, req.From)

	// set properties.
	if entity, err = s.entityManager.CreateEntity(ctx, entity); nil != err {
		log.Error("create entity failed",
			logger.EntityID(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "create entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "create entity failed")
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			entity.KValues[k] = constraint.NewNode(v)
		}
	case nil:
		log.Error("update entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityEmptyRequest))
		return nil, ErrEntityEmptyRequest
	default:
		log.Error("update entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityInvalidParams))
		return nil, ErrEntityInvalidParams
	}

	// check properties.
	if _, has := entity.KValues[""]; has {
		log.Error("update entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityPropertyIDEmpty))
		return out, ErrEntityPropertyIDEmpty
	}

	// set properties.
	if entity, err = s.entityManager.SetProperties(ctx, entity); nil != err {
		log.Error("update entity failed.", logger.EntityID(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "update entity failed")
}

func (s *EntityService) PatchEntity(ctx context.Context, req *pb.PatchEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	entity.KValues = make(map[string]constraint.Node)

	switch kv := req.Properties.AsInterface().(type) {
	case []interface{}:
		patchData := make([]*pb.PatchData, 0)
		data, _ := json.Marshal(kv)
		if err = json.Unmarshal(data, &patchData); nil != err {
			log.Error("patch entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityInvalidParams))
			return nil, ErrEntityInvalidParams
		}

		// check path data.
		for _, pd := range patchData {
			if err = checkPatchData(pd); nil != err {
				log.Error("patch entity failed.", logger.EntityID(req.Id), zap.Error(err))
				return nil, errors.Wrap(err, "patch entity failed")
			}
		}

		if entity, err = s.entityManager.PatchEntity(ctx, entity, patchData); nil != err {
			log.Error("patch entity failed.", logger.EntityID(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "patch entity failed")
		}
	case nil:
		log.Error("patch entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityEmptyRequest))
		return nil, ErrEntityEmptyRequest
	default:
		log.Error("patch entity failed.", logger.EntityID(req.Id), zap.Error(ErrEntityInvalidParams))
		return nil, ErrEntityInvalidParams
	}

	out = s.entity2EntityResponse(entity)
	return out, nil
}

func (s *EntityService) PatchEntityZ(ctx context.Context, req *pb.PatchEntityRequest) (out *pb.EntityResponse, err error) {
	return s.PatchEntity(ctx, req)
}

func checkPatchData(patchData *pb.PatchData) error {
	if constraint.IsReversedOp(patchData.Operator) {
		return constraint.ErrJSONPatchReservedOp
	} else if !constraint.IsValidPath(patchData.Path) {
		return constraint.ErrPatchPathInvalid
	}
	return nil
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (out *pb.DeleteEntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

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

func (s *EntityService) GetEntityProps(ctx context.Context, in *pb.GetEntityPropsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	pids := strings.Split(strings.TrimSpace(in.Pids), ",")
	if len(pids) == 0 {
		log.Error("get entity properties, empty property ids.", logger.EntityID(in.Id))
		return out, ErrEntityInvalidParams
	}

	// get entity from entity manager.
	if entity, err = s.entityManager.GetProperties(ctx, entity); nil != err {
		log.Error("get entity failed.", logger.EntityID(in.Id), zap.Error(err))
		return
	}

	props := make(map[string]constraint.Node)
	// patch copy.
	for _, pid := range pids {
		props[pid] = constraint.NewNode(nil)
		if !strings.ContainsAny(pid, ".[") {
			if val, exists := entity.KValues[pid]; exists {
				props[pid] = val
			}
			continue
		}

		arr := strings.SplitN(strings.TrimSpace(pid), ".", 2)
		// patch property.
		props[pid], err = constraint.Patch(entity.KValues[arr[0]], nil, arr[1], constraint.PatchOpCopy)
		if nil != err {
			log.Error("get entity failed.", logger.EntityID(in.Id), zap.Error(err))
			return
		}
	}

	entity.KValues = props
	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "get entity properties")
}

func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	// get entity from entity manager.
	if entity, err = s.entityManager.GetProperties(ctx, entity); nil != err {
		log.Error("get entity failed.", logger.EntityID(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "get entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return
}

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (out *pb.ListEntityResponse, err error) {
	searchReq := &pb.SearchRequest{}
	searchReq.Query = req.Query
	searchReq.Page = req.Page
	searchReq.Condition = req.Condition

	var resp *pb.SearchResponse
	if resp, err = s.searchClient.Search(ctx, searchReq); err != nil {
		log.Error("list entities failed.", zap.Error(err))
		return out, errors.Wrap(err, "list entity failed")
	}

	out = &pb.ListEntityResponse{}
	out.Total = resp.Total
	out.Limit = resp.Limit
	for _, item := range resp.Items {
		switch kv := item.AsInterface().(type) {
		case map[string]interface{}:
			properties, _ := structpb.NewValue(kv)
			entityItem := &pb.EntityResponse{
				Id:         interface2string(kv["id"]),
				Source:     req.Source,
				Owner:      req.Owner,
				Type:       "",
				Properties: properties,
				Mappers:    []*pb.MapperDesc{},
			}
			out.Items = append(out.Items, entityItem)
		}
	}
	if err != nil {
		log.Error("list entities failed", zap.Error(err))
		return out, errors.Wrap(err, "entity search failed")
	}
	return out, nil
}

func (s *EntityService) AppendMapper(ctx context.Context, req *pb.AppendMapperRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	mapperDesc := statem.MapperDesc{}
	if req.Mapper != nil {
		mapperDesc.Name = req.Mapper.Name
		mapperDesc.TQLString = req.Mapper.Tql
		entity.Mappers = []statem.MapperDesc{mapperDesc}
	} else {
		log.Error("append mapper failed.", logger.EntityID(req.Id), zap.Error(err))
		return nil, errors.Wrap(ErrEntityMapperNil, "append mapper to entity failed")
	}

	// set properties.
	entity, err = s.entityManager.AppendMapper(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2EntityResponse(entity)
	return
}

// SetConfigs set entity configs.
func (s *EntityService) SetConfigs(ctx context.Context, in *pb.SetConfigsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	if entity.Configs, err = parseConfigFrom(ctx, in.Configs.AsInterface()); nil != err {
		log.Error("set entity configs", logger.EntityID(in.Id), zap.Error(err))
		return out, err
	}

	// set properties.
	if entity, err = s.entityManager.SetConfigs(ctx, entity); nil != err {
		log.Error("set entity configs", logger.EntityID(in.Id), zap.Error(err))
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "set entity configs")
}

// AppendConfigs append entity configs.
func (s *EntityService) AppendConfigs(ctx context.Context, in *pb.AppendConfigsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	if entity.Configs, err = parseConfigFrom(ctx, in.Configs.AsInterface()); nil != err {
		log.Error("append entity configs", logger.EntityID(in.Id), zap.Error(err))
		return out, err
	}

	// set properties.
	if entity, err = s.entityManager.AppendConfigs(ctx, entity); nil != err {
		log.Error("append entity configs", logger.EntityID(in.Id), zap.Error(err))
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "append entity configs")
}

// RemoveConfigs remove entity configs.
func (s *EntityService) RemoveConfigs(ctx context.Context, in *pb.RemoveConfigsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	propertyIDs := strings.Split(in.PropertyIds, ",")
	if entity, err = s.entityManager.RemoveConfigs(ctx, entity, propertyIDs); nil != err {
		log.Error("remove entity configs", logger.EntityID(in.Id), zap.Error(err))
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "remove entity configs")
}

// QueryConfigs query entity configs.
func (s *EntityService) QueryConfigs(ctx context.Context, in *pb.QueryConfigsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	propertyIDs := strings.Split(in.PropertyIds, ",")
	if entity, err = s.entityManager.QueryConfigs(ctx, entity, propertyIDs); nil != err {
		log.Error("query entity configs", logger.EntityID(in.Id), zap.Error(err))
	}

	configs := make(map[string]constraint.Config)
	for key, cfg := range entity.Configs {
		configs[key] = cfg
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "query entity configs")
}

func (s *EntityService) PatchConfigs(ctx context.Context, in *pb.PatchConfigsRequest) (out *pb.EntityResponse, err error) {
	var entity = new(Entity)
	entity.ID = in.Id
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)
	entity.KValues = make(map[string]constraint.Node)

	switch kv := in.Configs.AsInterface().(type) {
	case []interface{}:
		patchData := make([]*pb.PatchData, 0)
		data, _ := json.Marshal(kv)
		if err = json.Unmarshal(data, &patchData); nil != err {
			log.Error("patch entity  configs", logger.EntityID(in.Id), zap.Error(ErrEntityInvalidParams))
			return nil, ErrEntityInvalidParams
		}

		var pds []*statem.PatchData
		for _, pd := range patchData {
			var cfg constraint.Config
			switch value := pd.Value.AsInterface().(type) {
			case map[string]interface{}:
				if cfg, err = constraint.ParseConfigsFrom(value); nil != err {
					return out, errors.Wrap(err, "parse entity configs")
				}
			}
			pds = append(pds, &statem.PatchData{Path: pd.Path, Operator: constraint.NewPatchOperator(pd.Operator), Value: cfg})
		}

		util.DebugInfo("decode patchdata", pds)
		if entity, err = s.entityManager.PatchConfigs(ctx, entity, pds); nil != err {
			log.Error("patch entity configs", logger.EntityID(in.Id), zap.Error(err))
			return nil, errors.Wrap(err, "patch entity failed")
		}
	case nil:
		log.Error("patch entity configs", logger.EntityID(in.Id), zap.Error(ErrEntityEmptyRequest))
		return nil, ErrEntityEmptyRequest
	default:
		log.Error("patch entity configs", logger.EntityID(in.Id), zap.Error(ErrEntityInvalidParams))
		return nil, ErrEntityInvalidParams
	}

	out = s.entity2EntityResponse(entity)
	return out, nil
}

// func (s *EntityService) PatchConfigsx(ctx context.Context, in *pb.PatchConfigsRequest) (out *pb.EntityResponse, err error) {
// 	var entity = new(Entity)
// 	entity.ID = in.Id
// 	entity.Owner = in.Owner
// 	entity.Source = in.Source
// 	parseHeaderFrom(ctx, entity)
// 	entity.KValues = make(map[string]constraint.Node)

// 	patchData := make([]*statem.PatchData, 0)

// 	for _, pd := range in.Data.Properties {
// 		operator := constraint.NewPatchOperator(pd.Operator)
// 		switch operator {
// 		case constraint.PatchOpAdd:
// 			fallthrough
// 		case constraint.PatchOpReplace:
// 			var cfgRet constraint.Config
// 			switch value := pd.Value.AsInterface().(type) {
// 			case map[string]interface{}:
// 				if cfgRet, err = constraint.ParseConfigsFrom(value); nil != err {
// 					return out, errors.Wrap(err, "parse entity config failed")
// 				}
// 			}
// 			patchData = append(patchData, &statem.PatchData{Path: pd.Path, Operator: operator, Value: cfgRet})
// 		case constraint.PatchOpRemove:
// 			fallthrough
// 		case constraint.PatchOpCopy:
// 			patchData = append(patchData, &statem.PatchData{Path: pd.Path, Operator: operator})
// 		case constraint.PatchOpUndef:
// 			log.Error("patch entity configs", zap.Error(constraint.ErrJSONPatchReservedOp), zap.String("op", pd.Operator))
// 			return out, constraint.ErrJSONPatchReservedOp
// 		}
// 	}

// 	util.DebugInfo("PatchData", patchData)

// 	entity, err = s.entityManager.PatchConfigs(ctx, entity, patchData)

// 	out = s.entity2EntityResponse(entity)
// 	return out, errors.Wrap(err, "patch entity configs")
// }

// parseConfigFrom parse config.
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
	case nil:
		log.Error("set entity configs failed.", zap.Error(ErrEntityEmptyRequest))
		return nil, ErrEntityEmptyRequest
	default:
		log.Error("set entity configs failed.", zap.Error(ErrEntityInvalidParams))
		return nil, ErrEntityConfigInvalid
	}
	return out, errors.Wrap(err, "parse entity config failed")
}

// parseHeaderFrom parse headers.
func parseHeaderFrom(ctx context.Context, en *statem.Base) {
	if header := ctx.Value(struct{}{}); nil != header {
		switch h := header.(type) {
		case http.Header:
			if en.Type == "" {
				en.Type = h.Get(HeaderType)
			}
			if en.Owner == "" {
				en.Owner = h.Get(HeaderOwner)
			}
			if en.Source == "" {
				en.Source = h.Get(HeaderSource)
			}
		default:
			panic("invalid HEADERS")
		}
	}
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
		log.Error("convert entity failed", zap.Error(err))
	} else if out.Configs, err = structpb.NewValue(configs); nil != err {
		log.Error("convert entity failed.", zap.Error(err))
	}

	out.Mappers = make([]*pb.MapperDesc, 0)
	for _, mapper := range entity.Mappers {
		out.Mappers = append(out.Mappers, &pb.MapperDesc{Name: mapper.Name, Tql: mapper.TQLString})
	}

	out.Source = entity.Source
	out.Owner = entity.Owner
	out.Id = entity.ID
	out.Type = entity.Type

	return out
}
