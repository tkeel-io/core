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
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/runtime/state"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/structpb"
)

type EntityService struct {
	pb.UnimplementedEntityServer

	inited       *atomic.Bool
	ctx          context.Context
	cancel       context.CancelFunc
	apiManager   apim.APIManager
	searchClient pb.SearchHTTPServer
}

func NewEntityService(ctx context.Context) (*EntityService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &EntityService{
		ctx:    ctx,
		cancel: cancel,
		inited: atomic.NewBool(false),
	}, nil
}

func (s *EntityService) Init(apiManager apim.APIManager, searchClient pb.SearchHTTPServer) {
	s.apiManager = apiManager
	s.searchClient = searchClient
	s.inited.Store(true)
}

func (s *EntityService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Type = req.Type
	entity.Source = req.Source
	entity.TemplateID = req.From
	parseHeaderFrom(ctx, entity)
	entity.Properties = make(map[string]tdtl.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		if entity.Properties, err = parseProps(kv); nil != err {
			log.Error("create entity, but invalid params",
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}
	case nil:
		log.Warn("create entity, but empty params", zfield.Eid(req.Id))
	default:
		log.Error("create entity, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, xerrors.ErrInvalidEntityParams
	}

	// check properties.
	if _, has := entity.Properties[""]; has {
		log.Error("create entity, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrEntityPropertyIDEmpty))
		return out, xerrors.ErrEntityPropertyIDEmpty
	}

	// set template entity id.
	ctx = context.WithValue(ctx, apim.TemplateEntityID{}, req.From)

	// set properties.
	if entity, err = s.apiManager.CreateEntity(ctx, entity); nil != err {
		log.Error("create entity failed", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "create entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "create entity failed")
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source

	parseHeaderFrom(ctx, entity)
	entity.Properties = make(map[string]tdtl.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		if entity.Properties, err = parseProps(kv); nil != err {
			log.Error("create entity, but invalid params",
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}
	case nil:
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	// check properties.
	if _, has := entity.Properties[""]; has {
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return out, xerrors.ErrInvalidRequest
	}

	// set properties.
	if entity, err = s.apiManager.UpdateEntityProps(ctx, entity); nil != err {
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "update entity failed")
}

func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	// get entity from entity manager.
	if entity, err = s.apiManager.GetEntity(ctx, entity); nil != err {
		log.Error("get entity failed.", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "get entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (out *pb.DeleteEntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	// delete entity.
	if err = s.apiManager.DeleteEntity(ctx, entity); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.ID(req.Id))
		return
	}

	out = &pb.DeleteEntityResponse{}
	out.Id = req.Id
	out.Status = "ok"
	return
}

func (s *EntityService) UpdateEntityProps(ctx context.Context, req *pb.UpdateEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source

	parseHeaderFrom(ctx, entity)
	entity.Properties = make(map[string]tdtl.Node)
	switch kv := req.Properties.AsInterface().(type) {
	case map[string]interface{}:
		if entity.Properties, err = parseProps(kv); nil != err {
			log.Error("create entity, but invalid params",
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}
	case nil:
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	// check properties.
	if _, has := entity.Properties[""]; has {
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return out, xerrors.ErrInvalidRequest
	}

	// set properties.
	if entity, err = s.apiManager.UpdateEntityProps(ctx, entity); nil != err {
		log.Error("update entity failed.", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity failed")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "update entity failed")
}

func (s *EntityService) PatchEntityProps(ctx context.Context, req *pb.PatchEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	entity.Properties = make(map[string]tdtl.Node)

	switch kv := req.Properties.AsInterface().(type) {
	case []interface{}:
		patchData := make([]PatchData, 0)
		if data, err := json.Marshal(kv); nil != err { //nolint
			log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json marshal patch data")
		} else if err = json.Unmarshal(data, &patchData); nil != err {
			log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json unmarshal patch data")
		}

		for index := range patchData {
			// encode value.
			bytes, _ := json.Marshal(patchData[index].Value)
			patchData[index].Value = bytes
		}

		pds := []state.PatchData{}
		for index := range patchData {
			pd := state.PatchData{
				Path:     patchData[index].Path,
				Operator: patchData[index].Operator,
				Value:    patchData[index].Value,
			}

			if err = checkPatchData(pd); nil != err {
				log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(err))
				return nil, errors.Wrap(err, "patch entity failed")
			}

			pds = append(pds, pd)
		}

		if entity, err = s.apiManager.PatchEntityProps(ctx, entity, pds); nil != err {
			log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "patch entity failed")
		}
	case nil:
		log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.Error("patch entity failed.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	out = s.entity2EntityResponse(entity)
	return out, nil
}

func (s *EntityService) PatchEntityPropsZ(ctx context.Context, req *pb.PatchEntityPropsRequest) (out *pb.EntityResponse, err error) {
	return s.PatchEntityProps(ctx, req)
}

func checkPatchData(patchData state.PatchData) error {
	if xjson.IsReversedOp(patchData.Operator) {
		return xerrors.ErrJSONPatchReservedOp
	} else if !xjson.IsValidPath(patchData.Path) {
		return xerrors.ErrPatchPathInvalid
	}
	return nil
}

func (s *EntityService) GetEntityProps(ctx context.Context, in *pb.GetEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	if pids := strings.Split(strings.TrimSpace(in.PropertyKeys), ","); len(pids) == 0 {
		log.Error("patch entity properties, empty property ids.", zfield.Eid(in.Id))
		return out, xerrors.ErrInvalidRequest
	}

	// get entity from entity manager.
	if entity, err = s.apiManager.GetEntity(ctx, entity); nil != err {
		log.Error("patch entity failed.", zfield.Eid(in.Id), zap.Error(err))
		return out, errors.Wrap(err, "get entity properties")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "get entity properties")
}

func (s *EntityService) RemoveEntityProps(ctx context.Context, in *pb.RemoveEntityPropsRequest) (*pb.EntityResponse, error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	panic("implement me")
}

// SetConfigs set entity configs.
func (s *EntityService) UpdateEntityConfigs(ctx context.Context, in *pb.UpdateEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	if entity.Configs, err = parseConfigFrom(ctx, in.Configs.AsInterface()); nil != err {
		log.Error("set entity configs", zfield.Eid(in.Id), zap.Error(err))
		return out, err
	}

	// set entity configs.
	if entity, err = s.apiManager.UpdateEntityConfigs(ctx, entity); nil != err {
		log.Error("set entity configs", zfield.Eid(in.Id), zap.Error(err))
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "set entity configs")
}

func (s *EntityService) PatchEntityConfigs(ctx context.Context, in *pb.PatchEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)
	entity.Properties = make(map[string]tdtl.Node)

	switch kv := in.Configs.AsInterface().(type) {
	case []interface{}:
		patchData := make([]*pb.PatchData, 0)
		data, _ := json.Marshal(kv)
		if err = json.Unmarshal(data, &patchData); nil != err {
			log.Error("patch entity  configs", zfield.Eid(in.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json unmarshal request")
		}

		var pds []state.PatchData
		for _, pd := range patchData {
			var cfg constraint.Config
			switch value := pd.Value.AsInterface().(type) {
			case map[string]interface{}:
				if cfg, err = constraint.ParseConfigFrom(value); nil != err {
					return out, errors.Wrap(err, "parse entity configs")
				}
			}

			var bytes []byte
			if bytes, err = json.Marshal(cfg); nil != err {
				log.Error("json marshal", zap.Error(err), zfield.Eid(in.Id))
				return nil, errors.Wrap(err, "patch entity failed")
			}
			pds = append(pds, state.PatchData{
				Path:     pd.Path,
				Operator: pd.Operator,
				Value:    bytes,
			})
		}

		if entity, err = s.apiManager.PatchEntityConfigs(ctx, entity, pds); nil != err {
			log.Error("patch entity configs", zfield.Eid(in.Id), zap.Error(err))
			return nil, errors.Wrap(err, "patch entity failed")
		}
	case nil:
		log.Error("patch entity configs", zfield.Eid(in.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.Error("patch entity configs", zfield.Eid(in.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	out = s.entity2EntityResponse(entity)
	return out, nil
}

func (s *EntityService) PatchEntityConfigsZ(ctx context.Context, req *pb.PatchEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	return s.PatchEntityConfigs(ctx, req)
}

// QueryConfigs query entity configs.
func (s *EntityService) GetEntityConfigs(ctx context.Context, in *pb.GetEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	var propertyIDs []string
	if in.PropertyKeys != "" {
		propertyIDs = strings.Split(strings.TrimSpace(in.PropertyKeys), ",")
	}

	if entity, err = s.apiManager.GetEntityConfigs(ctx, entity, propertyIDs); nil != err {
		log.Error("query entity configs", zfield.Eid(in.Id), zap.Error(err))
	}

	out = s.entity2EntityResponseZ(entity)
	return out, errors.Wrap(err, "query entity configs")
}

// RemoveConfigs remove entity configs.
func (s *EntityService) RemoveEntityConfigs(ctx context.Context, in *pb.RemoveEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	propertyIDs := strings.Split(in.PropertyKeys, ",")
	pds := make([]state.PatchData, 0)
	for index := range propertyIDs {
		pds = append(pds, state.PatchData{
			Path:     propertyIDs[index],
			Operator: xjson.OpRemove.String(),
		})
	}

	if entity, err = s.apiManager.PatchEntityConfigs(ctx, entity, pds); nil != err {
		log.Error("patch entity configs", zfield.Eid(in.Id), zap.Error(err))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	out = s.entity2EntityResponse(entity)
	return out, errors.Wrap(err, "remove entity configs")
}

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (out *pb.ListEntityResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Owner(req.Owner))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	searchReq := &pb.SearchRequest{}
	searchReq.Query = req.Query
	searchReq.PageNum = req.PageNum
	searchReq.PageSize = req.PageSize
	searchReq.IsDescending = req.IsDescending
	searchReq.OrderBy = req.OrderBy
	searchReq.Condition = req.Condition

	var resp *pb.SearchResponse
	if resp, err = s.searchClient.Search(ctx, searchReq); err != nil {
		log.Error("list apim failed.", zap.Error(err))
		return out, errors.Wrap(err, "list entity failed")
	}

	out = &pb.ListEntityResponse{}
	out.Total = resp.Total
	out.PageNum = resp.PageNum
	out.PageSize = resp.PageSize
	for _, item := range resp.Items {
		switch kv := item.AsInterface().(type) {
		case map[string]interface{}:
			id := interface2string(kv["id"])
			source := interface2string(kv["source"])
			owner := interface2string(kv["owner"])
			entityType := interface2string(kv["type"])

			delete(kv, "id")
			delete(kv, "source")
			delete(kv, "owner")
			delete(kv, "type")
			delete(kv, "version")
			delete(kv, "last_time")
			properties, _ := structpb.NewValue(kv)
			entityItem := &pb.EntityResponse{
				Id:         id,
				Source:     source,
				Owner:      owner,
				Type:       entityType,
				Properties: properties,
				Mappers:    []*pb.Mapper{},
			}
			out.Items = append(out.Items, entityItem)
		}
	}
	if err != nil {
		log.Error("list apim failed", zap.Error(err))
		return out, errors.Wrap(err, "entity search failed")
	}
	return out, nil
}

// parseConfigFrom parse config.
func parseConfigFrom(ctx context.Context, data interface{}) (out map[string]*constraint.Config, err error) {
	// parse configs from.
	out = make(map[string]*constraint.Config)
	switch configs := data.(type) {
	case []interface{}:
		for _, cfg := range configs {
			if c, ok := cfg.(map[string]interface{}); ok {
				var cfgRet constraint.Config
				if cfgRet, err = constraint.ParseConfigFrom(c); nil != err {
					return out, errors.Wrap(err, "parse entity config failed")
				}
				out[cfgRet.ID] = &cfgRet
				continue
			}
			return out, xerrors.ErrInvalidRequest
		}
	case nil:
		log.Error("set entity configs.", zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.Error("set entity configs.", zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}
	return out, errors.Wrap(err, "parse entity config")
}

// parseHeaderFrom parse headers.
func parseHeaderFrom(ctx context.Context, en *apim.Base) {
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

func (s *EntityService) entity2EntityResponseZ(entity *Entity) (out *pb.EntityResponse) {
	if entity == nil {
		return
	}

	var err error
	out = &pb.EntityResponse{}

	properties := make(map[string]interface{})
	bytes, _ := xjson.EncodeJSON(entity.Properties)
	json.Unmarshal(bytes, &properties)

	var configBytes []byte
	configs := make(map[string]interface{})
	if configBytes, err = json.Marshal(entity.Configs); nil != err {
		log.Error("marshal entity configs", zap.Error(err), zfield.Eid(entity.ID))
	} else if err = json.Unmarshal(configBytes, &configs); nil != err {
		log.Error("unmarshal entity configs", zap.Error(err), zfield.Eid(entity.ID))
	}

	if out.Properties, err = structpb.NewValue(properties); nil != err {
		log.Error("convert entity failed", zap.Error(err))
	} else if out.Configs, err = structpb.NewValue(configs); nil != err {
		log.Error("convert entity failed.", zap.Error(err))
	}

	out.Mappers = make([]*pb.Mapper, 0)
	for _, mDesc := range entity.Mappers {
		out.Mappers = append(out.Mappers,
			&pb.Mapper{
				Id:          mDesc.ID,
				Name:        mDesc.Name,
				TqlText:     mDesc.TQL,
				Description: mDesc.Description,
			})
	}

	out.Id = entity.ID
	out.Type = entity.Type
	out.Owner = entity.Owner
	out.Source = entity.Source

	return out
}

func (s *EntityService) entity2EntityResponse(entity *Entity) (out *pb.EntityResponse) {
	if entity == nil {
		return
	}

	var err error
	out = &pb.EntityResponse{}

	properties := make(map[string]interface{})
	bytes, _ := xjson.EncodeJSON(entity.Properties)
	json.Unmarshal(bytes, &properties)

	configs := make(map[string]interface{})
	if err = json.Unmarshal(entity.ConfigFile, &configs); nil != err {
		log.Error("unmarshal entity configs", zap.Error(err), zfield.Eid(entity.ID))
	}

	if out.Properties, err = structpb.NewValue(properties); nil != err {
		log.Error("convert entity failed", zap.Error(err))
	} else if out.Configs, err = structpb.NewValue(configs); nil != err {
		log.Error("convert entity failed.", zap.Error(err))
	}

	out.Mappers = make([]*pb.Mapper, 0)
	for _, mDesc := range entity.Mappers {
		out.Mappers = append(out.Mappers,
			&pb.Mapper{
				Id:          mDesc.ID,
				Name:        mDesc.Name,
				TqlText:     mDesc.TQL,
				Description: mDesc.Description,
			})
	}

	out.Id = entity.ID
	out.Type = entity.Type
	out.Owner = entity.Owner
	out.Source = entity.Source

	return out
}

func parseProps(props map[string]interface{}) (map[string]tdtl.Node, error) {
	bytes, err := json.Marshal(props)
	if nil != err {
		log.Error("marshal properties", zap.Error(err))
		return nil, errors.Wrap(err, "marshal properties")
	}

	var result = make(map[string]tdtl.Node)
	collectjs.ForEach(bytes, jsonparser.Object,
		func(key, value []byte, dataType jsonparser.ValueType) {
			result[string(key)] = xjson.NewNode(dataType, value)
		})

	return result, errors.Wrap(err, "parse properties")
}
