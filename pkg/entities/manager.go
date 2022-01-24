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

package entities

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities/proxy"
	zfield "github.com/tkeel-io/core/pkg/logger"
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

type entityManager struct {
	entityRepo   repository.IRepository
	stateManager state.Manager
	coreProxy    *proxy.Proxy

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(
	ctx context.Context,
	repo repository.IRepository,
	stateManager state.Manager) (EntityManager, error) {
	ctx, cancel := context.WithCancel(ctx)
	entityManager := &entityManager{
		ctx:          ctx,
		cancel:       cancel,
		entityRepo:   repo,
		stateManager: stateManager,
		lock:         sync.RWMutex{},
	}

	coreProxy, err := proxy.NewProxy(ctx, stateManager)
	if nil != err {
		log.Error("new Proxy instance", zap.Error(err))
		return nil, errors.Wrap(err, "new EntityManager")
	}

	entityManager.coreProxy = coreProxy
	return entityManager, nil
}

func (m *entityManager) Start() error {
	return errors.Wrap(m.stateManager.Start(), "start entity manager")
}

func (m *entityManager) OnMessage(ctx context.Context, msgCtx message.MessageContext) error {
	err := m.coreProxy.RouteMessage(ctx, msgCtx)
	return errors.Wrap(err, "core consume message")
}

// ------------------------------------APIs-----------------------------.

func (m *entityManager) checkID(base *Base) {
	if base.ID == "" {
		base.ID = util.UUID()
	}
}

// CreateEntity create a entity.
func (m *entityManager) CreateEntity(ctx context.Context, base *Base) (out *Base, err error) {
	var has bool
	var templateID string

	m.checkID(base)
	log.Info("entity.CreateEntity",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(base.JSON()))

	// 1. check entity exists.
	if has, err = m.entityRepo.HasEntity(ctx, &dao.Entity{ID: base.ID}); nil != err && has {
		log.Error("check entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	// 2. check template entity.
	if templateID, _ = ctx.Value(TemplateEntityID{}).(string); templateID != "" {
		has, err = m.entityRepo.HasEntity(ctx, &dao.Entity{ID: templateID})
		if nil != err && has {
			log.Error("check template", zap.Error(err), zfield.Eid(templateID))
			return nil, errors.Wrap(err, "create entity")
		}
	}

	waitG := sync.WaitGroup{}
	// send msg.
	waitG.Add(1)
	elapsedTime := util.NewElapsed()
	reqID, msgID := util.UUID(), util.UUID()
	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.FlushPropertyMessage{
			StateID:    base.ID,
			Operator:   constraint.PatchOpReplace.String(),
			Properties: base.Properties,
			MessageBase: message.NewBase(func(v interface{}) {
				waitG.Done()
				log.Debug("dispose message completed",
					zfield.Eid(base.ID), zfield.ReqID(reqID),
					zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
			}),
		},
	}

	msgCtx.Headers.SetType(base.Type)
	msgCtx.Headers.SetOwner(base.Owner)
	msgCtx.Headers.SetReceiver(base.ID)
	msgCtx.Headers.SetTemplate(templateID)
	msgCtx.Headers.SetRequestID(reqID)
	msgCtx.Headers.SetMessageID(msgID)
	msgCtx.Headers.SetSender(CoreAPISender)
	if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
		log.Error("create entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	waitG.Wait()
	return base, errors.Wrap(err, "create entity")
}

// DeleteEntity delete an entity from manager.
func (m *entityManager) DeleteEntity(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.DeleteEntity",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	waitG := sync.WaitGroup{}
	// send msg.
	waitG.Add(1)
	elapsedTime := util.NewElapsed()
	reqID, msgID := util.UUID(), util.UUID()
	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.StateMessage{
			StateID: base.ID,
			Method:  message.SMMethodDeleteEntity,
			MessageBase: message.NewBase(func(v interface{}) {
				waitG.Done()
				log.Debug("dispose message completed",
					zfield.Eid(base.ID), zfield.ReqID(reqID),
					zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
			}),
		},
	}

	msgCtx.Headers.SetType(base.Type)
	msgCtx.Headers.SetOwner(base.Owner)
	msgCtx.Headers.SetReceiver(base.ID)
	msgCtx.Headers.SetRequestID(reqID)
	msgCtx.Headers.SetMessageID(msgID)
	msgCtx.Headers.SetSender(CoreAPISender)
	if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "delete entity")
	}

	return base, errors.Wrap(err, "delete entity")
}

// GetProperties returns Base.
func (m *entityManager) GetProperties(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.GetProperties",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	var res *dao.Entity
	if res, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: en.ID}); nil != err {
		log.Error("get entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "get entity")
	}

	return convert(res), errors.Wrap(err, "get entity")
}

// SetProperties set properties into entity.
func (m *entityManager) SetProperties(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.SetProperties",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	waitG := sync.WaitGroup{}

	// send msg.
	waitG.Add(1)
	elapsedTime := util.NewElapsed()
	reqID, msgID := util.UUID(), util.UUID()
	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.FlushPropertyMessage{
			StateID:    base.ID,
			Operator:   constraint.PatchOpReplace.String(),
			Properties: base.Properties,
			MessageBase: message.NewBase(func(v interface{}) {
				waitG.Done()
				log.Debug("dispose message completed",
					zfield.Eid(base.ID), zfield.ReqID(reqID),
					zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
			}),
		},
	}

	msgCtx.Headers.SetType(base.Type)
	msgCtx.Headers.SetOwner(base.Owner)
	msgCtx.Headers.SetReceiver(base.ID)
	msgCtx.Headers.SetRequestID(reqID)
	msgCtx.Headers.SetMessageID(msgID)
	msgCtx.Headers.SetSender(CoreAPISender)
	if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
		log.Error("route entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "route entity")
	}

	waitG.Wait()
	return base, errors.Wrap(err, "set entity properties")
}

func (m *entityManager) PatchEntity(ctx context.Context, en *Base, patchData []*pb.PatchData) (base *Base, err error) {
	log.Info("entity.PatchEntity",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	// group by operator.
	pdm := make(map[string][]*pb.PatchData)
	for _, pd := range patchData {
		pdm[pd.Operator] = append(pdm[pd.Operator], pd)
	}

	reqID := util.UUID()
	waitG := sync.WaitGroup{}
	elapsedTime := util.NewElapsed()
	for op, pds := range pdm {
		kvs := make(map[string]constraint.Node)
		for _, pd := range pds {
			kvs[pd.Path] = constraint.NewNode(pd.Value.AsInterface())
		}

		if len(kvs) > 0 {
			waitG.Add(1)
			msgID := util.UUID()
			msgCtx := message.MessageContext{
				Headers: message.Header{},
				Message: message.FlushPropertyMessage{
					StateID:    en.ID,
					Operator:   op,
					Properties: kvs,
					MessageBase: message.NewBase(func(v interface{}) {
						waitG.Done()
						log.Debug("dispose message completed",
							zfield.Eid(base.ID), zfield.ReqID(reqID),
							zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
					}),
				},
			}

			// set headers.
			msgCtx.Headers.SetType(en.Type)
			msgCtx.Headers.SetOwner(en.Owner)
			msgCtx.Headers.SetReceiver(en.ID)
			msgCtx.Headers.SetSource(en.Source)
			msgCtx.Headers.SetSender(CoreAPISender)
			msgCtx.Headers.SetRequestID(reqID)
			msgCtx.Headers.SetMessageID(msgID)
			if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
				log.Error("route message", zfield.Eid(en.ID), zap.Error(err))
				return nil, errors.Wrap(err, "route message")
			}
		}
	}

	waitG.Wait()
	return base, errors.Wrap(err, "patch entity properties")
}

// AppendMapper append a mapper into entity.
func (m *entityManager) AppendMapper(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.AppendMapper",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	// upert mapper.
	mp := en.Mappers[0]
	err = m.entityRepo.PutMapper(ctx, &dao.Mapper{
		ID:          mp.ID,
		TQL:         mp.TQL,
		Name:        mp.Name,
		EntityID:    en.ID,
		EntityType:  en.Type,
		Description: mp.Description,
	})

	return base, errors.Wrap(err, "append mapper")
}

// DeleteMapper delete mapper from entity.
func (m *entityManager) RemoveMapper(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.RemoveMapper",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	// delete mapper.
	mp := en.Mappers[0]
	err = m.entityRepo.DelMapper(ctx, &dao.Mapper{
		ID:          mp.ID,
		TQL:         mp.TQL,
		Name:        mp.Name,
		EntityID:    en.ID,
		EntityType:  en.Type,
		Description: mp.Description,
	})

	return base, errors.Wrap(err, "remove mapper")
}

func (m *entityManager) CheckSubscription(ctx context.Context, en *Base) (err error) {
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
func (m *entityManager) SetConfigs(ctx context.Context, en *Base) (base *Base, err error) {
	log.Info("entity.SetConfigs",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	waitG := sync.WaitGroup{}
	// send msg.
	waitG.Add(1)
	elapsedTime := util.NewElapsed()
	reqID, msgID := util.UUID(), util.UUID()
	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.StateMessage{
			StateID: base.ID,
			Value:   en.Configs,
			Method:  message.SMMethodSetConfigs,
			MessageBase: message.NewBase(func(v interface{}) {
				waitG.Done()
				log.Debug("dispose message completed",
					zfield.Eid(base.ID), zfield.ReqID(reqID),
					zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
			}),
		},
	}

	msgCtx.Headers.SetType(base.Type)
	msgCtx.Headers.SetOwner(base.Owner)
	msgCtx.Headers.SetReceiver(base.ID)
	msgCtx.Headers.SetRequestID(reqID)
	msgCtx.Headers.SetMessageID(msgID)
	msgCtx.Headers.SetSender(CoreAPISender)
	if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
		log.Error("route entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "route entity")
	}

	waitG.Wait()
	return base, errors.Wrap(err, "set entity properties")
}

// PatchConfigs patch properties into entity.
func (m *entityManager) PatchConfigs(ctx context.Context, en *Base, patchData []*state.PatchData) (base *Base, err error) {
	log.Info("entity.PatchConfigs",
		zfield.Eid(en.ID), zfield.Type(en.Type),
		zfield.Owner(en.Owner), zfield.Source(en.Source), zfield.Base(en.JSON()))

	waitG := sync.WaitGroup{}
	// send msg.
	waitG.Add(1)
	elapsedTime := util.NewElapsed()
	reqID, msgID := util.UUID(), util.UUID()
	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.StateMessage{
			StateID: en.ID,
			Value:   patchData,
			Method:  message.SMMethodPatchConfigs,
			MessageBase: message.NewBase(func(v interface{}) {
				waitG.Done()
				log.Debug("dispose message completed",
					zfield.Eid(en.ID), zfield.ReqID(reqID),
					zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))
			}),
		},
	}

	msgCtx.Headers.SetType(en.Type)
	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetReceiver(en.ID)
	msgCtx.Headers.SetRequestID(reqID)
	msgCtx.Headers.SetMessageID(msgID)
	msgCtx.Headers.SetSender(CoreAPISender)
	if err = m.coreProxy.RouteMessage(ctx, msgCtx); nil != err {
		log.Error("route entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "route entity")
	}

	waitG.Wait()
	return base, nil
}

// QueryConfigs query entity configs.
func (m *entityManager) QueryConfigs(ctx context.Context, en *Base, propertyIDs []string) (base *Base, err error) {
	log.Info("entity.PatchConfigs",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	// get entity config file.

	// get properties by ids.

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
