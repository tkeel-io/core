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
	"fmt"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/mapper/tql"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const eventType = "Core.APIs"
const respondFmt = "http://%s:%d/v1/respond"

var msgTypeSync string = message.MessageTypeSync.String()

func eventSender(api string) string {
	return fmt.Sprintf("%s.%s", eventType, api)
}

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
		base.ID = util.UUID()
	}
}

func (m *apiManager) callbackAddr() string {
	return fmt.Sprintf(respondFmt, util.ResolveAddr(), config.Get().Proxy.HTTPPort)
}

// CreateEntity create a entity.
func (m *apiManager) CreateEntity(ctx context.Context, en *Base) (*Base, error) {
	var (
		err         error
		has         bool
		templateID  string
		elapsedTime = util.NewElapsed()
	)

	m.checkID(en)
	log.Info("entity.CreateEntity",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// 1. check entity exists.
	if has, err = m.entityRepo.HasEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("check entity exists", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "create entity")
	} else if has {
		log.Error("check entity exists", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(xerrors.ErrEntityAleadyExists, "create entity")
	}

	// 2. check template entity.
	if templateID, _ = ctx.Value(TemplateEntityID{}).(string); templateID != "" {
		en.TemplateID = templateID
		if has, err = m.entityRepo.HasEntity(ctx, &dao.Entity{ID: templateID}); nil != err && has {
			log.Error("check template entity", zap.Error(err), zfield.Eid(templateID))
			return nil, errors.Wrap(err, "create entity")
		}
	}

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, en.ID)
	ev.SetExtension(message.ExtEntityType, en.Type)
	ev.SetExtension(message.ExtEntityOwner, en.Owner)
	ev.SetExtension(message.ExtMessageReceiver, en.ID)
	ev.SetExtension(message.ExtEntitySource, en.Source)
	ev.SetExtension(message.ExtTemplateID, en.TemplateID)
	ev.SetExtension(message.ExtCallback, m.callbackAddr())
	ev.SetExtension(message.ExtMessageType, msgTypeSync)
	ev.SetExtension(message.ExtAPIIdentify, state.APICreateEntity.String())
	ev.SetExtension(message.ExtMessageSender, eventSender("CreateEntity"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// TODO: encode Request to event.Data.
	var bytes = []byte(`{}`)

	if err = ev.SetData(bytes); nil != err {
		log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	} else if err = ev.Validate(); nil != err {
		log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "validate event")
	}

	if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
		log.Error("create entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	log.Debug("processing message", zfield.Eid(en.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	resp := m.holder.Wait(ctx, eventID)
	if resp.Status != holder.StatusOK {
		log.Error("create entity",
			zap.Error(xerrors.New(resp.ErrCode)),
			zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, xerrors.New(resp.ErrCode)
	}

	// decode resp.Data.

	return nil, nil
}

// DeleteEntity delete an entity from manager.
func (m *apiManager) DeleteEntity(ctx context.Context, en *Base) error {
	elapsedTime := util.NewElapsed()
	log.Info("entity.DeleteEntity",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	ev := cloudevents.NewEvent()
	msgID, eventID := util.UUID(), util.UUID()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtEntityID, en.ID)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityType, en.Type)
	ev.SetExtension(message.ExtEntityOwner, en.Owner)
	ev.SetExtension(message.ExtMessageReceiver, en.ID)
	ev.SetExtension(message.ExtEntitySource, en.Source)
	ev.SetExtension(message.ExtCallback, m.callbackAddr())
	ev.SetExtension(message.ExtMessageSender, CoreAPISender)
	ev.SetExtension(message.ExtMessageType, msgTypeSync)
	ev.SetExtension(message.ExtAPIIdentify, state.APIDeleteEntity)
	ev.SetExtension(message.ExtMessageSender, eventSender("DeleteEntity"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	var err error
	if err = ev.Validate(); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(en.ID))
		return errors.Wrap(err, "delete entity")
	}

	// TODO: encode Request to event.Data.
	var bytes []byte

	if err = ev.SetData(bytes); nil != err {
		log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return errors.Wrap(err, "encode props message")
	} else if err = ev.Validate(); nil != err {
		log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return errors.Wrap(err, "validate event")
	}

	if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(en.ID))
		return errors.Wrap(err, "delete entity")
	}

	log.Debug("processing message completed", zfield.Eid(en.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	resp := m.holder.Wait(ctx, eventID)
	if resp.Status != holder.StatusOK {
		log.Error("create entity",
			zap.Error(xerrors.New(resp.ErrCode)),
			zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return xerrors.New(resp.ErrCode)
	}

	// decode resp.Data.
	return nil
}

// GetProperties returns Base.
func (m *apiManager) GetProperties(ctx context.Context, en *Base) (*Base, error) {
	log.Info("entity.GetProperties",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	var err error
	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("get entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "get entity")
	}

	return entityToBase(entity), errors.Wrap(err, "get entity")
}

// SetProperties set properties into entity.
func (m *apiManager) SetProperties(ctx context.Context, en *Base) (*Base, error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.SetProperties", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, en.ID)
	ev.SetExtension(message.ExtEntityType, en.Type)
	ev.SetExtension(message.ExtEntityOwner, en.Owner)
	ev.SetExtension(message.ExtMessageReceiver, en.ID)
	ev.SetExtension(message.ExtEntitySource, en.Source)
	ev.SetExtension(message.ExtCallback, m.callbackAddr())
	ev.SetExtension(message.ExtMessageType, msgTypeSync)
	ev.SetExtension(message.ExtAPIIdentify, state.APISetProperties)
	ev.SetExtension(message.ExtMessageSender, eventSender("SetProperties"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// TODO: encode Request to event.Data.
	var bytes []byte
	var err error

	if err = ev.SetData(bytes); nil != err {
		log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	} else if err = ev.Validate(); nil != err {
		log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "validate event")
	}

	if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
		log.Error("set entity properties", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "set entity properties")
	}

	log.Debug("process message completed", zfield.Eid(en.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("set entity properties", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "set entity properties")
	}

	return entityToBase(entity), errors.Wrap(err, "set entity properties")
}

func (m *apiManager) PatchEntity(ctx context.Context, en *Base, patchData []*pb.PatchData) (*Base, error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.PatchEntity",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// group by operator.
	pdm := make(map[string][]*pb.PatchData)
	for _, pd := range patchData {
		pdm[pd.Operator] = append(pdm[pd.Operator], pd)
	}

	var err error
	reqID := util.UUID()
	for _, pds := range pdm {
		kvs := make(map[string]constraint.Node)
		for _, pd := range pds {
			kvs[pd.Path] = constraint.NewNode(pd.Value.AsInterface())
		}

		if len(kvs) > 0 {
			msgID := util.UUID()
			eventID := util.UUID()
			ev := cloudevents.NewEvent()
			ev.SetID(eventID)
			ev.SetType(eventType)
			ev.SetSource(config.Get().Server.Name)
			ev.SetExtension(message.ExtMessageID, msgID)
			ev.SetExtension(message.ExtEntityID, en.ID)
			ev.SetExtension(message.ExtEntityType, en.Type)
			ev.SetExtension(message.ExtEntityOwner, en.Owner)
			ev.SetExtension(message.ExtMessageReceiver, en.ID)
			ev.SetExtension(message.ExtEntitySource, en.Source)
			ev.SetExtension(message.ExtCallback, m.callbackAddr())
			ev.SetExtension(message.ExtMessageType, msgTypeSync)
			ev.SetExtension(message.ExtAPIIdentify, state.APIPatchEntity)
			ev.SetExtension(message.ExtMessageSender, eventSender("PatchEntity"))
			ev.SetDataContentType(cloudevents.ApplicationJSON)

			// TODO: encode Request to event.Data.
			var bytes []byte

			if err = ev.SetData(bytes); nil != err {
				log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
				return nil, errors.Wrap(err, "encode props message")
			} else if err = ev.Validate(); nil != err {
				log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
				return nil, errors.Wrap(err, "validate event")
			}

			if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
				log.Error("patch entity", zap.Error(err), zfield.Eid(en.ID))
				return nil, errors.Wrap(err, "patch entity")
			}

			log.Debug("processing message completed", zfield.Eid(en.ID),
				zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))
		}
	}

	log.Debug("processing message completed", zfield.Eid(en.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("patch entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "patch entity")
	}

	return entityToBase(entity), errors.Wrap(err, "patch entity")
}

// AppendMapper append a mapper into entity.
func (m *apiManager) AppendMapper(ctx context.Context, en *Base) (*Base, error) {
	log.Info("entity.AppendMapper",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// upert mapper.
	var err error
	mp := en.Mappers[0]
	if err = m.entityRepo.PutMapper(ctx, &dao.Mapper{
		ID:          mp.ID,
		TQL:         mp.TQL,
		Name:        mp.Name,
		EntityID:    en.ID,
		EntityType:  en.Type,
		Description: mp.Description,
	}); nil != err {
		log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "append mapper")
	}

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "append mapper")
	}

	return entityToBase(entity), errors.Wrap(err, "append mapper")
}

// DeleteMapper delete mapper from entity.
func (m *apiManager) RemoveMapper(ctx context.Context, en *Base) (*Base, error) {
	log.Info("entity.RemoveMapper",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	// delete mapper.
	var err error
	mp := en.Mappers[0]
	if err = m.entityRepo.DelMapper(ctx, &dao.Mapper{
		ID:          mp.ID,
		TQL:         mp.TQL,
		Name:        mp.Name,
		EntityID:    en.ID,
		EntityType:  en.Type,
		Description: mp.Description,
	}); nil != err {
		log.Error("remove mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "remove mapper")
	}

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("remove mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "remove mapper")
	}

	return entityToBase(entity), errors.Wrap(err, "remove mapper")
}

func (m *apiManager) CheckSubscription(ctx context.Context, en *Base) (err error) {
	// check TQLs.
	if err = checkTQLs(en); nil != err {
		return errors.Wrap(err, "check subscription")
	}

	// check request.
	mode := getString(en.Properties[subscription.SubscriptionFieldMode])
	topic := getString(en.Properties[subscription.SubscriptionFieldTopic])
	filter := getString(en.Properties[subscription.SubscriptionFieldFilter])
	pubsubName := getString(en.Properties[subscription.SubscriptionFieldPubsubName])
	log.Infof("check subscription, mode: %s, topic: %s, filter:%s, pubsub: %s, source: %s", mode, topic, filter, pubsubName, en.Source)
	if mode == subscription.SubscriptionModeUndefine || en.Source == "" || filter == "" || topic == "" || pubsubName == "" {
		log.Error("create subscription", zap.Error(runtime.ErrSubscriptionInvalid), zap.String("subscription", en.ID))
		return runtime.ErrSubscriptionInvalid
	}

	return nil
}

// SetProperties set properties into entity.
func (m *apiManager) SetConfigs(ctx context.Context, en *Base) (*Base, error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.SetConfigs", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, en.ID)
	ev.SetExtension(message.ExtEntityType, en.Type)
	ev.SetExtension(message.ExtEntityOwner, en.Owner)
	ev.SetExtension(message.ExtMessageReceiver, en.ID)
	ev.SetExtension(message.ExtEntitySource, en.Source)
	ev.SetExtension(message.ExtCallback, m.callbackAddr())
	ev.SetExtension(message.ExtMessageType, msgTypeSync)
	ev.SetExtension(message.ExtAPIIdentify, state.APISetConfigs)
	ev.SetExtension(message.ExtMessageSender, eventSender("SetConfigs"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// TODO: encode Request to event.Data.
	var bytes []byte
	var err error

	if err = ev.SetData(bytes); nil != err {
		log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	} else if err = ev.Validate(); nil != err {
		log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "validate event")
	}

	if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
		log.Error("set entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "set entity configs")
	}

	log.Debug("processing message completed", zfield.Eid(en.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("set entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "set entity configs")
	}

	return entityToBase(entity), errors.Wrap(err, "set entity configs")
}

// PatchConfigs patch properties into entity.
func (m *apiManager) PatchConfigs(ctx context.Context, en *Base, patchData []*state.PatchData) (*Base, error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.PatchConfigs", zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, en.ID)
	ev.SetExtension(message.ExtEntityType, en.Type)
	ev.SetExtension(message.ExtEntityOwner, en.Owner)
	ev.SetExtension(message.ExtMessageReceiver, en.ID)
	ev.SetExtension(message.ExtEntitySource, en.Source)
	ev.SetExtension(message.ExtCallback, m.callbackAddr())
	ev.SetExtension(message.ExtMessageType, msgTypeSync)
	ev.SetExtension(message.ExtAPIIdentify, state.APIPatchConfigs)
	ev.SetExtension(message.ExtMessageSender, eventSender("PatchConfigs"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// TODO: encode Request to event.Data.
	var bytes []byte
	var err error

	if err = ev.SetData(bytes); nil != err {
		log.Error("encode props message", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	} else if err = ev.Validate(); nil != err {
		log.Error("validate event", zap.Error(err), zfield.Eid(en.ID), zfield.Base(en.JSON()))
		return nil, errors.Wrap(err, "validate event")
	}

	if err = m.dispatcher.Dispatch(ctx, ev); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	// debug informations.
	log.Debug("processing message completed", zfield.Eid(en.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	return entityToBase(entity), errors.Wrap(err, "patch entity configs")
}

// QueryConfigs query entity configs.
func (m *apiManager) QueryConfigs(ctx context.Context, en *Base, propertyIDs []string) (*Base, error) {
	log.Info("entity.PatchConfigs",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	var err error
	var entity *dao.Entity
	// get entity config file.
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	// get properties by ids.
	cc := collectjs.ByteNew(entity.ConfigFile)
	configs := make(map[string]*constraint.Config)
	for _, propertyID := range propertyIDs {
		var cfg *constraint.Config
		bytes := cc.Get(propertyID).GetRaw()
		if cfg, err = constraint.ParseFrom(bytes); nil != err {
			log.Error("parse entity configs", zap.Error(err), zfield.Eid(en.ID))
			return nil, errors.Wrap(err, "parse entity configs")
		}
		configs[propertyID] = cfg
	}

	base := entityToBase(entity)
	base.Configs = configs

	return base, nil
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
		var tqlInst tql.TQL
		if tqlInst, err = tql.NewTQL(mm.TQL); nil != err {
			log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID))
			return errors.Wrap(err, "check TQL")
		} else if tqlInst.Target() != en.ID {
			log.Error("mismatched subscription id & mapper target id.", zfield.Eid(en.ID), zap.Any("mapper", mm))
			return errors.Wrap(err, "subscription ID mismatched")
		}
	}
	return errors.Wrap(err, "check TQL")
}

func getString(node constraint.Node) string {
	if nil != node {
		return node.String()
	}
	return ""
}
