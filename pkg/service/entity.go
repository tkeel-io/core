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
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/scheme"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	sep              = "."
	FieldScheme      = "scheme"
	FieldProps       = "properties"
	FieldTemplate    = "template_id"
	FieldDescription = "description"
	InternalSep      = ".define.fields."
)

func schemeKey(key string) string {
	key = strings.ReplaceAll(key, sep, InternalSep)
	return strings.Join([]string{FieldScheme, key}, sep)
}

func rawSchemeKey(key string) string {
	return strings.ReplaceAll(key, InternalSep, sep)
}

func propKey(key string) string {
	return strings.Join([]string{FieldProps, key}, sep)
}

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
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Type = req.Type
	entity.Source = req.Source
	entity.TemplateID = req.From
	parseHeaderFrom(ctx, entity)
	properties := req.Properties.AsInterface()
	switch properties.(type) {
	case map[string]interface{}:
		if entity.Properties, err = json.Marshal(properties); nil != err {
			log.L().Error("create entity, invalid params", zfield.Reason(err.Error()),
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}
	case nil:
		log.L().Warn("create entity, empty params", zfield.Eid(req.Id))
	default:
		log.L().Error("create entity, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, xerrors.ErrInvalidEntityParams
	}

	var baseRet *apim.BaseRet
	if baseRet, err = s.apiManager.CreateEntity(ctx, entity); nil != err {
		log.L().Error("create entity failed", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "create entity failed")
	}

	// ignore error.
	s.onTemplateChanged(ctx, entity)

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "create entity failed")
}

func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	patches := []*pb.PatchData{}

	log.L().Debug("update entity",
		zfield.Eid(req.Id), zfield.Owner(entity.Owner),
		zfield.Template(req.TemplateId), zfield.Desc(req.Description),
		zap.Any("scheme", req.Configs), zap.Any("properties", req.Properties))

	properties := req.Properties.AsInterface()
	switch properties.(type) {
	case map[string]interface{}:
		if entity.Properties, err = json.Marshal(properties); nil != err {
			log.L().Error("create entity, invalid params", zfield.Reason(err.Error()),
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}

		// patch merge properties.
		if len(entity.Properties) > 0 {
			patches = append(patches, &pb.PatchData{
				Path:     FieldProps,
				Value:    entity.Properties,
				Operator: xjson.OpMerge.String()})
		}
	case nil:
	default:
		log.L().Error("update entity failed.",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	if template := strings.TrimSpace(req.TemplateId); len(template) > 0 {
		entity.TemplateID = template
		patches = append(patches, &pb.PatchData{
			Path:     FieldTemplate,
			Value:    tdtl.NewString(template).Raw(),
			Operator: xjson.OpReplace.String(),
		})
	}

	if len(req.Description) > 0 {
		patches = append(patches, &pb.PatchData{
			Path:     FieldDescription,
			Value:    tdtl.NewString(req.Description).Raw(),
			Operator: xjson.OpReplace.String(),
		})
	}

	schemeVal := req.Configs.AsInterface()
	switch schemeVal.(type) {
	case map[string]interface{}:
		if entity.Scheme, err = json.Marshal(schemeVal); nil != err {
			log.L().Error("update entity, invalid params", zfield.Reason(err.Error()),
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "update entity")
		}

		// patch replace configs.
		if len(entity.Scheme) > 0 {
			patches = append(patches, &pb.PatchData{
				Path:     FieldScheme,
				Value:    entity.Scheme,
				Operator: xjson.OpReplace.String()})
		}
	case nil:
	default:
		log.L().Error("update entity failed.",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	var baseRet *apim.BaseRet
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("update entity failed.", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity failed")
	}

	// ignore error.
	s.onTemplateChanged(ctx, entity)
	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "update entity failed")
}

func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	var baseRet *apim.BaseRet
	if baseRet, err = s.apiManager.GetEntity(ctx, entity); nil != err {
		log.L().Error("get entity", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "get entity")
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "get entity")
}

func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (out *pb.DeleteEntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	// delete entity.
	if err = s.apiManager.DeleteEntity(ctx, entity); nil != err {
		log.L().Error("delete entity", zap.Error(err), zfield.ID(req.Id))
		return nil, errors.Wrap(err, "delete entity")
	}

	return &pb.DeleteEntityResponse{Id: req.Id, Status: "ok"}, nil
}

func (s *EntityService) UpdateEntityProps(ctx context.Context, req *pb.UpdateEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	properties := req.Properties.AsInterface()
	switch properties.(type) {
	case map[string]interface{}:
		if entity.Properties, err = json.Marshal(properties); nil != err {
			log.L().Error("create entity, but invalid params",
				zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
			return out, errors.Wrap(err, "create entity")
		}
	case nil:
		log.L().Error("update entity failed.",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.L().Error("update entity failed.",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	patches := []*pb.PatchData{{
		Path:     FieldProps,
		Operator: xjson.OpMerge.String(),
		Value:    entity.Properties,
	}}

	var baseRet *apim.BaseRet
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("update entity properties.", zfield.Eid(req.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity properties")
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "update entity properties")
}

func (s *EntityService) PatchEntityProps(ctx context.Context, req *pb.PatchEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = req.Id
	entity.Type = req.Type
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	patches := []*pb.PatchData{}
	params := req.Properties.AsInterface()
	switch params.(type) {
	case []interface{}:
		var data []byte
		patchData := make([]PatchData, 0)
		if data, err = json.Marshal(params); nil != err {
			log.L().Error("patch entity properties.", zfield.Eid(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json marshal patch data")
		} else if err = json.Unmarshal(data, &patchData); nil != err {
			log.L().Error("patch entity properties.", zfield.Eid(req.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json unmarshal patch data")
		}

		for index := range patchData {
			var bytes []byte
			if err = checkPatchData(patchData[index]); nil != err {
				log.L().Error("patch entity properties.", zfield.Eid(req.Id), zap.Error(err))
				return nil, errors.Wrap(err, "patch entity properties")
			} else if bytes, err = json.Marshal(patchData[index].Value); nil != err {
				return nil, errors.Wrap(err, "encode property")
			}
			// encode value.
			patches = append(patches, &pb.PatchData{
				Path:     propKey(patchData[index].Path),
				Operator: patchData[index].Operator,
				Value:    bytes,
			})
		}
	default:
		log.L().Error("patch entity properties.", zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	var rawEntity []byte
	var baseRet *apim.BaseRet
	if baseRet, rawEntity, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("patch entity properties.", zfield.Eid(req.Id), zap.Error(err))
		return nil, errors.Wrap(err, "patch entity properties")
	}

	// clip copy properties.
	if properties, cpflag, innerErr := CopyFrom(rawEntity, patches...); nil != innerErr {
		log.L().Warn("patch entity properties.", zfield.Eid(req.Id), zfield.Reason(err.Error()))
	} else if cpflag {
		baseRet.Properties = properties
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "patch entity properties")
}

func (s *EntityService) PatchEntityPropsZ(ctx context.Context, req *pb.PatchEntityPropsRequest) (out *pb.EntityResponse, err error) {
	return s.PatchEntityProps(ctx, req)
}

func checkPatchData(patchData PatchData) error {
	if xjson.IsReversedOp(patchData.Operator) {
		return xerrors.ErrJSONPatchReservedOp
	} else if !xjson.IsValidPath(patchData.Path) {
		return xerrors.ErrPatchPathInvalid
	}
	return nil
}

func (s *EntityService) GetEntityProps(ctx context.Context, in *pb.GetEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	var propKeys []string
	if pidsStr := strings.TrimSpace(in.PropertyKeys); len(pidsStr) > 0 {
		for _, key := range strings.Split(pidsStr, ",") {
			propKeys = append(propKeys, propKey(key))
		}
	}

	var rawEntity []byte
	var baseRet *apim.BaseRet
	// get entity from entity manager.
	if baseRet, rawEntity, err = s.apiManager.PatchEntity(ctx, entity, nil); nil != err {
		log.L().Error("patch entity failed.", zfield.Eid(in.Id), zap.Error(err))
		return out, errors.Wrap(err, "get entity properties")
	}

	// clip copy properties.
	if props, cpflag, innerErr := CopyFrom2(rawEntity, propKeys...); nil != innerErr {
		log.L().Warn("patch entity properties.", zfield.Eid(in.Id), zfield.Reason(innerErr.Error()))
	} else if cpflag {
		baseRet.Properties = props
	}

	baseRet.Scheme = nil
	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "get entity properties")
}

func (s *EntityService) RemoveEntityProps(ctx context.Context, in *pb.RemoveEntityPropsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	var propertyKeys []string
	if propertyKeys = strings.Split(strings.TrimSpace(in.PropertyKeys), ","); len(propertyKeys) == 0 {
		log.L().Error("remove entity properties, empty property ids.", zfield.Eid(in.Id))
		return out, xerrors.ErrInvalidRequest
	}

	patches := make([]*pb.PatchData, 0)
	for index := range propertyKeys {
		patches = append(patches, &pb.PatchData{
			Path:     propKey(propertyKeys[index]),
			Operator: xjson.OpRemove.String(),
		})
	}

	var baseRet *apim.BaseRet
	// get entity from entity manager.
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("patch entity failed.", zfield.Eid(in.Id), zap.Error(err))
		return out, errors.Wrap(err, "remove entity properties")
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "remove entity properties")
}

// SetConfigs set entity configs.
func (s *EntityService) UpdateEntityConfigs(ctx context.Context, in *pb.UpdateEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := &Entity{}
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)
	param := in.Configs.AsInterface()
	switch param.(type) {
	// TODO: 这里在后面调整 API 的时候换成 map[string]interfae{}.
	case []interface{}:
		var configs map[string]*scheme.Config
		if configs, err = parseSchemeFrom(param); nil != err {
			log.L().Error("update entity scheme", zfield.Eid(in.Id), zap.Error(err))
			return out, err
		} else if entity.Scheme, err = json.Marshal(configs); nil != err {
			log.L().Error("encode entity scheme", zfield.Eid(in.Id), zap.Error(err))
			return out, errors.Wrap(err, "encode scheme")
		}
	default:
		log.L().Error("update entity scheme.",
			zfield.Eid(in.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	patches := []*pb.PatchData{{
		Path:     FieldScheme,
		Operator: xjson.OpMerge.String(),
		Value:    entity.Scheme,
	}}

	// set entity configs.
	var baseRet *apim.BaseRet
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("update entity scheme", zfield.Eid(in.Id), zap.Error(err))
		return out, errors.Wrap(err, "update entity scheme")
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "update entity scheme")
}

func (s *EntityService) PatchEntityConfigs(ctx context.Context, in *pb.PatchEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	var patches []*pb.PatchData
	param := in.Configs.AsInterface()
	switch patchDatas := param.(type) {
	case []interface{}:
		patchData := make([]PatchData, 0)
		data, _ := json.Marshal(patchDatas)
		if err = json.Unmarshal(data, &patchData); nil != err {
			log.L().Error("patch entity scheme", zfield.Eid(in.Id), zap.Error(err))
			return nil, errors.Wrap(err, "json unmarshal request")
		}

		for index := range patchData {
			if err = checkPatchData(patchData[index]); nil != err {
				log.L().Error("check entity scheme.", zfield.Eid(in.Id), zap.Error(err))
				return nil, errors.Wrap(err, "patch entity scheme")
			}

			var cfg scheme.Config
			switch value := patchData[index].Value.(type) {
			case map[string]interface{}:
				if cfg, err = scheme.ParseConfigFrom(value); nil != err {
					log.L().Error("check entity scheme.", zfield.Eid(in.Id), zap.Error(err))
					return out, errors.Wrap(err, "parse entity scheme")
				}
			}

			var bytes []byte
			if bytes, err = json.Marshal(cfg); nil != err {
				log.L().Error("json marshal", zap.Error(err), zfield.Eid(in.Id))
				return nil, errors.Wrap(err, "patch entity scheme")
			}
			patches = append(patches, &pb.PatchData{
				Path:     schemeKey(patchData[index].Path),
				Operator: patchData[index].Operator,
				Value:    bytes,
			})
		}

	case nil:
		log.L().Error("patch entity scheme", zfield.Eid(in.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.L().Error("patch entity scheme", zfield.Eid(in.Id), zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	}

	var rawEntity []byte
	var baseRet *apim.BaseRet
	opts := []apim.Option{apim.NewPathConstructorOption(pb.PCScheme)}
	if baseRet, rawEntity, err = s.apiManager.PatchEntity(ctx, entity, patches, opts...); nil != err {
		log.L().Error("patch entity scheme", zfield.Eid(in.Id), zap.Error(err))
		return nil, errors.Wrap(err, "patch entity scheme")
	}

	// clip copy scheme.
	if scheme, cpflag, innerErr := CopyFrom(rawEntity, patches...); nil != innerErr {
		log.L().Warn("patch entity scheme.", zfield.Eid(in.Id), zfield.Reason(innerErr.Error()))
	} else if cpflag {
		baseRet.Scheme = make(map[string]interface{})
		for path, schemeValue := range scheme {
			rawPath := rawSchemeKey(path)
			baseRet.Scheme[rawPath] = schemeValue
		}
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "patch entity scheme")
}

func (s *EntityService) PatchEntityConfigsZ(ctx context.Context, req *pb.PatchEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	return s.PatchEntityConfigs(ctx, req)
}

// QueryConfigs query entity configs.
func (s *EntityService) GetEntityConfigs(ctx context.Context, in *pb.GetEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	var propKeys []string
	if pidStr := strings.TrimSpace(in.PropertyKeys); len(pidStr) > 0 {
		for _, key := range strings.Split(pidStr, ",") {
			propKeys = append(propKeys, schemeKey(key))
		}
	}

	var rawEntity []byte
	var baseRet *apim.BaseRet
	if baseRet, rawEntity, err = s.apiManager.PatchEntity(ctx, entity, nil); nil != err {
		log.L().Error("query entity scheme", zfield.Eid(in.Id), zap.Error(err))
		return nil, errors.Wrap(err, "get entity scheme")
	}

	if len(propKeys) > 0 {
		// clip copy properties.
		if scheme, cpflag, innerErr := CopyFrom2(rawEntity, propKeys...); nil != innerErr {
			log.L().Warn("patch entity scheme.", zfield.Eid(in.Id), zfield.Reason(innerErr.Error()))
		} else if cpflag {
			baseRet.Scheme = make(map[string]interface{})
			for path, schemeValue := range scheme {
				rawPath := rawSchemeKey(path)
				baseRet.Scheme[rawPath] = schemeValue
			}
		}
	}

	baseRet.Properties = nil
	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "get entity scheme")
}

// RemoveConfigs remove entity configs.
func (s *EntityService) RemoveEntityConfigs(ctx context.Context, in *pb.RemoveEntityConfigsRequest) (out *pb.EntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(in.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	entity := new(Entity)
	entity.ID = in.Id
	entity.Type = in.Type
	entity.Owner = in.Owner
	entity.Source = in.Source
	parseHeaderFrom(ctx, entity)

	// set properties.
	propertyIDs := strings.Split(in.PropertyKeys, ",")
	patches := make([]*pb.PatchData, 0)
	for index := range propertyIDs {
		patches = append(patches, &pb.PatchData{
			Path:     schemeKey(propertyIDs[index]),
			Operator: xjson.OpRemove.String(),
		})
	}

	var baseRet *apim.BaseRet
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("patch entity scheme", zfield.Eid(in.Id), zap.Error(err))
		return nil, errors.Wrap(err, "patch entity scheme")
	}

	out, err = s.makeResponse(baseRet)
	return out, errors.Wrap(err, "remove entity scheme")
}

func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (out *pb.ListEntityResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Owner(req.Owner))
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
		log.L().Error("list entity.", zap.Error(err))
		return out, errors.Wrap(err, "list entity")
	}

	out = &pb.ListEntityResponse{}
	out.Total = int32(resp.Total)
	out.PageNum = resp.PageNum
	out.PageSize = resp.PageSize
	for _, item := range resp.Items {
		switch kv := item.AsInterface().(type) {
		case map[string]interface{}:
			var entity = new(Entity)
			entity.ID = interface2string(kv["id"])
			entity.Source = interface2string(kv["source"])
			entity.Owner = interface2string(kv["owner"])
			entity.Type = interface2string(kv["type"])

			var baseRet *apim.BaseRet
			if baseRet, err = s.apiManager.GetEntity(ctx, entity); nil != err {
				log.L().Error("get entity failed.", zfield.Eid(interface2string(kv["id"])), zap.Error(err))
				continue
			}

			entityItem, _ := s.makeResponse(baseRet)
			out.Items = append(out.Items, entityItem)
		}
	}

	if err != nil {
		log.L().Error("list apim failed", zap.Error(err))
		// return out, errors.Wrap(err, "entity search failed")
	}
	return out, nil
}

// parseSchemeFrom parse config.
func parseSchemeFrom(data interface{}) (out map[string]*scheme.Config, err error) {
	// parse configs from.
	out = make(map[string]*scheme.Config)
	switch configs := data.(type) {
	case []interface{}:
		for _, cfg := range configs {
			if c, ok := cfg.(map[string]interface{}); ok {
				var cfgRet scheme.Config
				if cfgRet, err = scheme.ParseConfigFrom(c); nil != err {
					return out, errors.Wrap(err, "parse entity config failed")
				}
				out[cfgRet.ID] = &cfgRet
				continue
			}
			return out, xerrors.ErrInvalidRequest
		}
	case nil:
		log.L().Error("set entity configs.", zap.Error(xerrors.ErrInvalidRequest))
		return nil, xerrors.ErrInvalidRequest
	default:
		log.L().Error("set entity configs.", zap.Error(xerrors.ErrInvalidRequest))
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

func (s *EntityService) makeResponse(base *apim.BaseRet) (out *pb.EntityResponse, err error) {
	if base == nil {
		return
	}

	out = &pb.EntityResponse{}
	if out.Properties, err = structpb.NewValue(base.Properties); nil != err {
		log.L().Error("convert entity properties", zap.Error(err), zfield.ID(base.ID))
		return out, errors.Wrap(err, "convert entity properties")
	} else if out.Configs, err = structpb.NewValue(base.Scheme); nil != err {
		log.L().Error("convert entity scheme.", zap.Error(err), zfield.ID(base.ID))
		return out, errors.Wrap(err, "convert entity scheme")
	}

	out.Mappers = make([]*pb.Mapper, 0)
	for _, mDesc := range base.Mappers {
		out.Mappers = append(out.Mappers,
			&pb.Mapper{
				Id:          mDesc.Id,
				Name:        mDesc.Name,
				Tql:         mDesc.Tql,
				Description: mDesc.Description,
			})
	}

	out.Id = base.ID
	out.Type = base.Type
	out.Owner = base.Owner
	out.Source = base.Source
	out.Version = base.Version
	out.LastTime = base.LastTime
	out.TemplateId = base.TemplateID
	out.Description = base.Description
	return out, nil
}

func CopyFrom(raw []byte, patches ...*pb.PatchData) (map[string]interface{}, bool, error) {
	var cpFlag bool
	cc := tdtl.New(raw)
	result := make(map[string]interface{})
	for _, patch := range patches {
		switch patch.Operator {
		case xjson.OpCopy.String():
			cpFlag = true
			var val interface{}
			if ret := cc.Get(patch.Path); ret.Error() != nil {
				return nil, false, errors.Wrap(ret.Error(), "clip result")
			} else if err := json.Unmarshal(ret.Raw(), &val); nil != err {
				return nil, false, errors.Wrap(err, "clip result")
			}

			index := strings.IndexByte(patch.Path, '.')
			result[patch.Path[index+1:]] = val
		default:
		}
	}
	return result, cpFlag, nil
}

func CopyFrom2(raw []byte, paths ...string) (map[string]interface{}, bool, error) {
	patches := []*pb.PatchData{}
	for _, path := range paths {
		patches = append(patches, &pb.PatchData{
			Path:     path,
			Operator: xjson.OpCopy.String(),
		})
	}
	result, flag, err := CopyFrom(raw, patches...)
	return result, flag, errors.Wrap(err, "copy result")
}

func (s *EntityService) onTemplateChanged(ctx context.Context, en *Entity) error {
	if en.TemplateID == "" {
		return nil
	}

	// create a mapper for sync scheme.
	// insert into eid select template.scheme as scheme
	mp := &mapper.Mapper{
		ID:          "SyncScheme",
		TQL:         fmt.Sprintf("insert into %s select %s.scheme as scheme", en.ID, en.TemplateID),
		Name:        "SyncScheme",
		Owner:       en.Owner,
		EntityID:    en.ID,
		Description: "mapper instance to sync scheme",
	}

	log.L().Info("onTemplateChanged", zfield.Eid(en.ID), zfield.ID(mp.ID), zfield.TQL(mp.TQL))

	if err := s.apiManager.AppendMapperZ(ctx, mp); nil != err {
		log.L().Error("create template mapper", zap.Error(err), zfield.Eid(en.ID))
		return errors.Wrap(err, "create template mapper")
	}

	return nil
}
