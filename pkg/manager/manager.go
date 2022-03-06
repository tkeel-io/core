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
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

const evIDPrefix = "ev"
const reqIDPrefix = "req"
const eventSender = "Core.APIManager"
const respondFmt = "http://%s:%d/v1/respond"
const (
	sysET = string(v1.ETSystem)
	enET  = string(v1.ETEntity)
)

type apiManager struct {
	holder     holder.Holder
	dispatcher dispatch.Dispatcher
	entityRepo repository.IRepository

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func New(
	ctx context.Context,
	repo repository.IRepository,
	dispatcher dispatch.Dispatcher) (APIManager, error) {
	ctx, cancel := context.WithCancel(ctx)
	apiManager := &apiManager{
		ctx:        ctx,
		cancel:     cancel,
		holder:     holder.New(),
		entityRepo: repo,
		dispatcher: dispatcher,
		lock:       sync.RWMutex{},
	}

	return apiManager, nil
}

func (m *apiManager) Start() error {
	log.Info("start API Manager")
	return nil
}

func (m *apiManager) OnRespond(ctx context.Context, resp *holder.Response) {
	m.holder.OnRespond(resp)
}

// ------------------------------------APIs-----------------------------.

func (m *apiManager) checkID(base *Base) {
	if base.ID == "" {
		base.ID = util.UUID("en")
	}
}

func (m *apiManager) callbackAddr() string {
	return fmt.Sprintf(respondFmt, util.ResolveAddr(), config.Get().Proxy.HTTPPort)
}

// CreateEntity create a entity.
func (m *apiManager) CreateEntity(ctx context.Context, en *Base) (*BaseRet, error) {
	var (
		err        error
		has        bool
		bytes      []byte
		templateID string
	)

	m.checkID(en)
	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.CreateEntity", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// 2. check template entity.
	if templateID, _ = ctx.Value(TemplateEntityID{}).(string); templateID != "" {
		if has, err = m.entityRepo.HasEntity(ctx, &dao.Entity{ID: templateID}); nil != err {
			log.Error("check template entity", zap.Error(err), zfield.Eid(templateID), zfield.ReqID(reqID))
			return nil, errors.Wrap(err, "create entity")
		} else if !has {
			log.Error("check template entity", zfield.Eid(en.ID), zfield.ReqID(reqID),
				zap.Error(xerrors.ErrTemplateNotFound), zfield.Template(templateID))
			return nil, errors.Wrap(xerrors.ErrTemplateNotFound, "create entity")
		}
	}

	if bytes, err = dao.GetEntityCodec().Encode(&dao.Entity{
		ID:         en.ID,
		Type:       en.Type,
		Owner:      en.Owner,
		Source:     en.Source,
		TemplateID: templateID,
		Properties: en.Properties,
	}); nil != err {
		log.Error("create entity", zfield.Eid(en.ID), zfield.Type(en.Type),
			zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity")
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
		Id:        util.UUID(evIDPrefix),
		Timestamp: time.Now().UnixNano(),
		Callback:  m.callbackAddr(),
		Metadata: map[string]string{
			v1.MetaType:      sysET,
			v1.MetaRequestID: reqID,
			v1.MetaEntityID:  en.ID,
		},
		Data: &v1.ProtoEvent_SystemData{
			SystemData: &v1.SystemData{
				Operator: string(v1.OpCreate),
				Data:     bytes,
			},
		},
	}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("create entity", zfield.Eid(en.ID), zfield.ReqID(reqID),
			zap.Error(xerrors.New(resp.ErrCode)), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if err = m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response, list mapper")
	}

	return &baseRet, errors.Wrap(err, "create entity")
}

func (m *apiManager) UpdateEntity(ctx context.Context, en *Base) (*BaseRet, error) {
	var (
		err   error
		bytes []byte
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.UpdateEntity", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))

	if bytes, err = xjson.EncodeJSON(en.Properties); nil != err {
		log.Error("update entity", zfield.Eid(en.ID), zfield.Type(en.Type),
			zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))
		return nil, errors.Wrap(err, "update entity")
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{{
						Path:     "properties",
						Value:    bytes,
						Operator: string(runtime.OpMerge),
					}},
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("update entity", zfield.Eid(en.ID),
			zfield.ReqID(reqID), zap.Error(xerrors.New(resp.ErrCode)))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if err = m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response, list mapper")
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// GetProperties returns Base.
func (m *apiManager) GetEntity(ctx context.Context, en *Base) (*BaseRet, error) {
	var err error
	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.GetProperties", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{{
						Path:     "",
						Operator: string(runtime.OpCopy),
					}},
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("get entity", zfield.Eid(en.ID),
			zfield.ReqID(reqID), zap.Error(xerrors.New(resp.ErrCode)))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// DeleteEntity delete an entity from manager.
func (m *apiManager) DeleteEntity(ctx context.Context, en *Base) error {
	var err error

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.DeleteEntity", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
		Id:        util.UUID(evIDPrefix),
		Timestamp: time.Now().UnixNano(),
		Callback:  m.callbackAddr(),
		Metadata: map[string]string{
			v1.MetaType:      sysET,
			v1.MetaRequestID: reqID,
			v1.MetaEntityID:  en.ID,
		},
		Data: &v1.ProtoEvent_SystemData{
			SystemData: &v1.SystemData{
				Operator: string(v1.OpDelete),
			},
		},
	}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("delete entity", zfield.Eid(en.ID),
			zfield.ReqID(reqID), zap.Error(xerrors.New(resp.ErrCode)))
		return xerrors.New(resp.ErrCode)
	}

	// decode response.
	var apiResp dao.Entity
	if err = dao.GetEntityCodec().Decode(resp.Data, &apiResp); nil != err {
		log.Error("delete entity, decode response",
			zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return errors.Wrap(err, "delete entity, decode response")
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return nil
}

// SetProperties set properties into entity.
func (m *apiManager) UpdateEntityProps(ctx context.Context, en *Base) (*BaseRet, error) {
	var (
		err   error
		bytes []byte
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.UpdateEntityProps", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	if bytes, err = xjson.EncodeJSON(en.Properties); nil != err {
		log.Error("update entity", zfield.Eid(en.ID), zfield.Type(en.Type),
			zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))
		return nil, errors.Wrap(err, "update entity")
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{{
						Path:     "properties",
						Value:    bytes,
						Operator: string(runtime.OpMerge),
					}},
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("set entity properties", zap.Error(xerrors.New(resp.ErrCode)),
			zfield.Eid(en.ID), zfield.ReqID(reqID), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

func (m *apiManager) PatchEntityProps(ctx context.Context, en *Base, pds []*v1.PatchData) (*BaseRet, error) {
	var (
		err error
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.PatchEntity", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: pds,
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("patch entity properties", zfield.Eid(en.ID),
			zap.Error(xerrors.New(resp.ErrCode)), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

func (m *apiManager) GetEntityProps(ctx context.Context, en *Base, propertyKeys []string) (*BaseRet, error) {
	var (
		err error
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.GetEntityProps", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))

	patches := make([]*v1.PatchData, 0)
	if len(propertyKeys) > 0 {
		for _, propKey := range propertyKeys {
			patches = append(patches,
				&v1.PatchData{
					Path:     "properties." + propKey,
					Operator: string(runtime.OpCopy),
				})
		}
	} else {
		patches = append(patches, &v1.PatchData{
			Path:     "properties",
			Operator: string(runtime.OpCopy),
		})
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patches,
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("get entity props", zfield.Eid(en.ID),
			zfield.ReqID(reqID), zap.Error(xerrors.New(resp.ErrCode)))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// SetProperties set properties into entity.
func (m *apiManager) UpdateEntityConfigs(ctx context.Context, en *Base) (*BaseRet, error) {
	var (
		err   error
		bytes []byte
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.UpdateEntityConfigs", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	if bytes, err = json.Marshal(en.Configs); nil != err {
		log.Error("json marshal configs", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "encode entity configs")
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
				v1.MetaType:      enET,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{{
						Path:     "scheme",
						Value:    bytes,
						Operator: string(runtime.OpMerge),
					}},
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("set entity configs", zfield.Eid(en.ID), zfield.ReqID(reqID),
			zap.Error(xerrors.New(resp.ErrCode)), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// PatchConfigs patch properties into entity.
func (m *apiManager) PatchEntityConfigs(ctx context.Context, en *Base, pds []*v1.PatchData) (*BaseRet, error) {
	var (
		err error
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.PatchConfigs", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: pds,
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("set entity configs", zap.Error(xerrors.New(resp.ErrCode)),
			zfield.Eid(en.ID), zfield.ReqID(reqID), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	// decode response.
	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// QueryConfigs query entity configs.
func (m *apiManager) GetEntityConfigs(ctx context.Context, en *Base, propertyKeys []string) (*BaseRet, error) {
	var (
		err error
	)

	reqID := util.UUID(reqIDPrefix)
	elapsedTime := util.NewElapsed()
	log.Info("entity.GetEntityConfigs", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.ReqID(reqID), zfield.Owner(en.Owner), zfield.Source(en.Source))

	patches := make([]*v1.PatchData, 0)
	if len(propertyKeys) > 0 {
		for _, propKey := range propertyKeys {
			patches = append(patches,
				&v1.PatchData{
					Path:     "scheme." + propKey,
					Operator: string(runtime.OpCopy),
				})
		}
	} else {
		patches = append(patches, &v1.PatchData{
			Path:     "scheme",
			Operator: string(runtime.OpCopy),
		})
	}

	// dispatch event.
	if err = m.dispatcher.Dispatch(ctx,
		&v1.ProtoEvent{
			Id:        util.UUID(evIDPrefix),
			Timestamp: time.Now().UnixNano(),
			Callback:  m.callbackAddr(),
			Metadata: map[string]string{
				v1.MetaType:      enET,
				v1.MetaRequestID: reqID,
				v1.MetaEntityID:  en.ID,
			},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patches,
				},
			},
		}); nil != err {
		log.Error("create entity, dispatch event", zap.Error(err), zfield.Eid(en.ID), zfield.ReqID(reqID))
		return nil, errors.Wrap(err, "create entity, dispatch event")
	}

	log.Debug("holding request, wait response", zfield.Eid(en.ID), zfield.ReqID(reqID))

	// hold request, wait response.
	resp := m.holder.Wait(ctx, reqID)
	if resp.Status != types.StatusOK {
		log.Error("get entity configs", zfield.Eid(en.ID),
			zfield.ReqID(reqID), zap.Error(xerrors.New(resp.ErrCode)))
		return nil, xerrors.New(resp.ErrCode)
	}

	// decode response.
	var baseRet BaseRet
	if err = json.Unmarshal(resp.Data, &baseRet); nil != err {
		log.Error("create entity, decode response", zfield.ReqID(reqID),
			zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "create entity, decode response")
	} else if innerErr := m.addMapper(ctx, &baseRet); nil != err {
		log.Error("create entity, decode response, list mapper", zfield.ReqID(reqID),
			zap.Error(innerErr), zfield.Eid(en.ID), zfield.Base(en.JSON()))
	}

	log.Info("processing completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	return &baseRet, errors.Wrap(err, "update entity")
}

// AppendMapper append a mapper into entity.
func (m *apiManager) AppendMapper(ctx context.Context, mp *dao.Mapper) error {
	log.Info("entity.AppendMapper",
		zfield.ID(mp.ID), zfield.Eid(mp.EntityID), zfield.Owner(mp.Owner))

	var err error
	// upert mapper.
	if err = m.entityRepo.PutMapper(ctx, mp); nil != err {
		log.Error("append mapper", zap.Error(err), zfield.ID(mp.ID), zfield.Eid(mp.EntityID))
		return errors.Wrap(err, "append mapper")
	}

	return nil
}

// DeleteMapper delete mapper from entity.
func (m *apiManager) RemoveMapper(ctx context.Context, mp *dao.Mapper) error {
	log.Info("entity.RemoveMapper",
		zfield.ID(mp.ID), zfield.Eid(mp.EntityID), zfield.Owner(mp.Owner))

	// delete mapper.
	var err error
	if err = m.entityRepo.DelMapper(ctx, mp); nil != err {
		log.Error("remove mapper", zap.Error(err), zfield.ID(mp.ID), zfield.Eid(mp.EntityID))
		return errors.Wrap(err, "remove mapper")
	}

	return nil
}

func (m *apiManager) GetMapper(ctx context.Context, mp *dao.Mapper) (*dao.Mapper, error) {
	log.Info("entity.GetMapper",
		zfield.ID(mp.ID), zfield.Eid(mp.EntityID), zfield.Owner(mp.Owner))

	// delete mapper.
	var err error
	if mp, err = m.entityRepo.GetMapper(ctx, mp); nil != err {
		log.Error("get mapper", zap.Error(err), zfield.ID(mp.ID), zfield.Eid(mp.EntityID))
		return mp, errors.Wrap(err, "get mapper")
	}

	return mp, nil
}

func (m *apiManager) ListMapper(ctx context.Context, en *Base) ([]dao.Mapper, error) {
	log.Info("entity.GetMapper", zfield.Eid(en.ID), zfield.Owner(en.Owner))

	// delete mapper.
	var err error
	var mps []dao.Mapper
	if mps, err = m.entityRepo.ListMapper(ctx,
		m.entityRepo.GetLastRevision(ctx),
		&dao.ListMapperReq{
			Owner:    en.Owner,
			EntityID: en.ID,
		}); nil != err {
		log.Error("list mapper", zap.Error(err), zfield.Eid(en.ID), zfield.Owner(en.Owner))
		return mps, errors.Wrap(err, "list mapper")
	}

	return mps, nil
}

func (m *apiManager) addMapper(ctx context.Context, base *BaseRet) error {
	mappers, err := m.entityRepo.ListMapper(ctx,
		m.entityRepo.GetLastRevision(ctx),
		&dao.ListMapperReq{
			Owner:    base.Owner,
			EntityID: base.ID,
		})
	if nil != err {
		return errors.Wrap(err, "list mapper by entity id.")
	}

	for _, mp := range mappers {
		base.Mappers = append(base.Mappers,
			&v1.Mapper{
				Id:          mp.ID,
				Tql:         mp.TQL,
				Name:        mp.Name,
				Description: mp.Description,
			})
	}

	return nil
}

func checkTQLs(en *Base) error {
	// check TQL.
	var err error
	defer func() {
		defer func() {
			switch recover() {
			case nil:
			default:
				err = ErrMapperTQLInvalid
			}
		}()
	}()
	for _, mm := range en.Mappers {
		var tqlInst tdtl.TDTL
		if tqlInst, err = tdtl.NewTDTL(mm.Tql, nil); nil != err {
			log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID))
			return errors.Wrap(err, "check TQL")
		} else if tqlInst.Target() != en.ID {
			log.Error("mismatched subscription id & mapper target id.", zfield.Eid(en.ID), zap.Any("mapper", mm))
			return errors.Wrap(err, "subscription ID mismatched")
		}
	}
	return errors.Wrap(err, "check TQL")
}

func getString(node tdtl.Node) string {
	if nil != node {
		return node.String()
	}
	return ""
}
