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
	"fmt"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities/proxy"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper/tql"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const eventType = "Core.APIs"

func eventSender(api string) string {
	return fmt.Sprintf("%s.%s", eventType, api)
}

type entityManager struct {
	entityRepo   repository.IRepository
	stateManager state.Manager
	coreProxy    *proxy.Proxy
	receivers    map[string]pubsub.Receiver

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
		receivers:    make(map[string]pubsub.Receiver),
		stateManager: stateManager,
		lock:         sync.RWMutex{},
	}

	// set state manager Republisher.
	stateManager.SetRepublisher(entityManager.coreProxy)

	coreProxy, err := proxy.NewProxy(ctx, stateManager)
	if nil != err {
		log.Error("new Proxy instance", zap.Error(err))
		return nil, errors.Wrap(err, "new EntityManager")
	}

	entityManager.coreProxy = coreProxy
	return entityManager, nil
}

func (m *entityManager) listQueue() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()
	revision := m.entityRepo.GetLastRevision(ctx)
	coreNodeName := config.Get().Server.Name
	m.entityRepo.RangeQueue(context.Background(), revision, func(queues []dao.Queue) {
		// create receiver.
		for _, queue := range queues {
			if coreNodeName == queue.NodeName {
				log.Info("append queue", zfield.ID(queue.ID))
				// create receiver instance.
				receiver := pubsub.NewPubsub(resource.Metadata{
					Name:       queue.Type.String(),
					Properties: queue.Metadata,
				})

				if _, has := m.receivers[queue.ID]; has {
					m.receivers[queue.ID].Close()
				}
				m.receivers[queue.ID] = receiver
			}
		}
	})
}

func (m *entityManager) watchQueue() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()
	revision := m.entityRepo.GetLastRevision(ctx)

	coreNodeName := config.Get().Server.Name
	ctx, cancel1 := context.WithCancel(m.ctx)
	defer cancel1()
	m.entityRepo.WatchQueue(ctx, revision, func(et dao.EnventType, queue dao.Queue) {
		switch et {
		case dao.PUT:
			// create receiver.
			if coreNodeName == queue.NodeName {
				log.Info("upsert queue", zfield.ID(queue.ID))
				// create receiver instance.
				receiver := pubsub.NewPubsub(resource.Metadata{
					Name:       queue.Type.String(),
					Properties: queue.Metadata,
				})

				if _, has := m.receivers[queue.ID]; has {
					m.receivers[queue.ID].Close()
				}
				m.receivers[queue.ID] = receiver
				// start consumer queue.
				receiver.Received(context.Background(), func(ctx context.Context, ev cloudevents.Event) error {
					if err := m.coreProxy.RouteMessage(ctx, ev); nil != err {
						// TODO: 对出处理错误的消息，需要做出处理.
						log.Error("route event", zap.Error(err), zap.String("queue", queue.ID), zfield.Event(ev))
					}
					log.Debug("received event", zap.String("queue", queue.ID), zfield.Event(ev))
					return nil
				})
			}
		case dao.DELETE:
			log.Info("remove queue", zfield.ID(queue.ID))
			if _, has := m.receivers[queue.ID]; has {
				log.Error("catch Queue event", zfield.ID(queue.ID),
					zfield.Type(queue.Type.String()), zfield.Name(queue.Name))
			}
		default:
		}
	})
}

func (m *entityManager) Start() error {
	// start runtime.
	if err := m.stateManager.Start(); nil != err {
		log.Error("start state manager")
		return errors.Wrap(err, "start state manager")
	}

	m.listQueue()
	go m.watchQueue()
	for id, receiver := range m.receivers {
		receiver.Received(context.Background(), func(ctx context.Context, ev cloudevents.Event) error {
			if err := m.coreProxy.RouteMessage(ctx, ev); nil != err {
				// TODO: 对出处理错误的消息，需要做出处理.
				log.Error("route event", zap.Error(err), zap.String("queue", id), zfield.Event(ev))
			}
			log.Debug("received event", zap.String("queue", id), zfield.Event(ev))
			return nil
		})
	}

	return nil
}

func (m *entityManager) OnMessage(ctx context.Context, e cloudevents.Event) error {
	err := m.coreProxy.RouteMessage(ctx, e)
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
	var (
		has         bool
		templateID  string
		elapsedTime = util.NewElapsed()
	)

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

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, base.ID)
	ev.SetExtension(message.ExtEntityType, base.Type)
	ev.SetExtension(message.ExtSyncFlag, message.Sync)
	ev.SetExtension(message.ExtEntityOwner, base.Owner)
	ev.SetExtension(message.ExtMessageReceiver, base.ID)
	ev.SetExtension(message.ExtEntitySource, base.Source)
	ev.SetExtension(message.ExtTemplateID, base.TemplateID)
	ev.SetExtension(message.ExtMessageType, message.MessageTypeProps)
	ev.SetExtension(message.ExtMessageSender, eventSender("CreateEntity"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// encode message.
	bytes, err := message.GetPropsCodec().Encode(message.PropertyMessage{
		StateID:    base.ID,
		Properties: base.Properties,
		Operator:   constraint.PatchOpReplace.String(),
	})

	if nil != err {
		log.Error("encode props message", zap.Error(err),
			zfield.Eid(base.ID), zfield.Base(base.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	}

	ev.SetData(bytes)

	if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
		log.Error("create entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	log.Debug("process message completed", zfield.Eid(base.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("create entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	return entityToBase(entity), errors.Wrap(err, "create entity")
}

// DeleteEntity delete an entity from manager.
func (m *entityManager) DeleteEntity(ctx context.Context, en *Base) (base *Base, err error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.DeleteEntity",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, base.ID)
	ev.SetExtension(message.ExtEntityType, base.Type)
	ev.SetExtension(message.ExtSyncFlag, message.Sync)
	ev.SetExtension(message.ExtEntityOwner, base.Owner)
	ev.SetExtension(message.ExtEntitySource, base.Source)
	ev.SetExtension(message.ExtMessageReceiver, base.ID)
	ev.SetExtension(message.ExtMessageSender, CoreAPISender)
	ev.SetExtension(message.ExtMessageType, message.MessageTypeState)
	ev.SetExtension(message.ExtMessageSender, eventSender("DeleteEntity"))

	// encode message.
	ev.SetData(message.StateMessage{
		StateID: base.ID,
		Method:  message.SMMethodDeleteEntity,
	})

	if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "delete entity")
	}

	log.Debug("dispose message completed", zfield.Eid(base.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "delete entity")
	}

	return entityToBase(entity), errors.Wrap(err, "delete entity")
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

	return entityToBase(res), errors.Wrap(err, "get entity")
}

// SetProperties set properties into entity.
func (m *entityManager) SetProperties(ctx context.Context, en *Base) (base *Base, err error) {
	elapsedTime := util.NewElapsed()
	log.Info("entity.SetProperties", zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, base.ID)
	ev.SetExtension(message.ExtEntityType, base.Type)
	ev.SetExtension(message.ExtSyncFlag, message.Sync)
	ev.SetExtension(message.ExtEntityOwner, base.Owner)
	ev.SetExtension(message.ExtMessageReceiver, base.ID)
	ev.SetExtension(message.ExtEntitySource, base.Source)
	ev.SetExtension(message.ExtMessageType, message.MessageTypeProps)
	ev.SetExtension(message.ExtMessageSender, eventSender("SetProperties"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// encode message.
	bytes, err := message.GetPropsCodec().Encode(message.PropertyMessage{
		StateID:    base.ID,
		Properties: base.Properties,
		Operator:   constraint.PatchOpReplace.String(),
	})

	if nil != err {
		log.Error("encode props message", zap.Error(err),
			zfield.Eid(base.ID), zfield.Base(base.JSON()))
		return nil, errors.Wrap(err, "encode props message")
	}

	ev.SetData(bytes)

	if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
		log.Error("set entity properties", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "set entity properties")
	}

	log.Debug("process message completed", zfield.Eid(base.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("set entity properties", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "set entity properties")
	}

	return entityToBase(entity), errors.Wrap(err, "set entity properties")
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
	elapsedTime := util.NewElapsed()
	for op, pds := range pdm {
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
			ev.SetExtension(message.ExtEntityID, base.ID)
			ev.SetExtension(message.ExtEntityType, base.Type)
			ev.SetExtension(message.ExtSyncFlag, message.Sync)
			ev.SetExtension(message.ExtEntityOwner, base.Owner)
			ev.SetExtension(message.ExtMessageReceiver, base.ID)
			ev.SetExtension(message.ExtEntitySource, base.Source)
			ev.SetExtension(message.ExtMessageType, message.MessageTypeProps)
			ev.SetExtension(message.ExtMessageSender, eventSender("PatchEntity"))
			ev.SetDataContentType(cloudevents.ApplicationJSON)

			// encode message.
			var bytes []byte
			bytes, err = message.GetPropsCodec().Encode(message.PropertyMessage{
				Operator:   op,
				StateID:    base.ID,
				Properties: base.Properties,
			})

			if nil != err {
				log.Error("encode props message", zap.Error(err),
					zfield.Eid(base.ID), zfield.Base(base.JSON()))
				return nil, errors.Wrap(err, "encode props message")
			}

			ev.SetData(bytes)

			if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
				log.Error("patch entity", zap.Error(err), zfield.Eid(base.ID))
				return nil, errors.Wrap(err, "patch entity")
			}
		}
	}

	log.Debug("dispose message completed", zfield.Eid(base.ID),
		zfield.ReqID(reqID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("patch entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "patch entity")
	}

	return entityToBase(entity), errors.Wrap(err, "patch entity")
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
	elapsedTime := util.NewElapsed()
	log.Info("entity.SetConfigs", zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, base.ID)
	ev.SetExtension(message.ExtEntityType, base.Type)
	ev.SetExtension(message.ExtSyncFlag, message.Sync)
	ev.SetExtension(message.ExtEntityOwner, base.Owner)
	ev.SetExtension(message.ExtMessageReceiver, base.ID)
	ev.SetExtension(message.ExtEntitySource, base.Source)
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetExtension(message.ExtMessageType, message.MessageTypeState)
	ev.SetExtension(message.ExtMessageSender, eventSender("SetConfigs"))

	// encode message.
	ev.SetData(message.StateMessage{
		StateID: base.ID,
		Value:   en.Configs,
		Method:  message.SMMethodSetConfigs,
	})

	if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
		log.Error("set entity configs", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "set entity configs")
	}

	log.Debug("dispose message completed", zfield.Eid(base.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("set entity configs", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "set entity configs")
	}

	return entityToBase(entity), errors.Wrap(err, "set entity configs")
}

// PatchConfigs patch properties into entity.
func (m *entityManager) PatchConfigs(ctx context.Context, en *Base, patchData []*state.PatchData) (base *Base, err error) {
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
	ev.SetExtension(message.ExtEntityID, base.ID)
	ev.SetExtension(message.ExtEntityType, base.Type)
	ev.SetExtension(message.ExtSyncFlag, message.Sync)
	ev.SetExtension(message.ExtEntityOwner, base.Owner)
	ev.SetExtension(message.ExtMessageReceiver, base.ID)
	ev.SetExtension(message.ExtEntitySource, base.Source)
	ev.SetExtension(message.ExtMessageType, message.MessageTypeState)
	ev.SetExtension(message.ExtMessageSender, eventSender("PatchConfigs"))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	// encode message.
	ev.SetData(message.StateMessage{
		StateID: en.ID,
		Value:   patchData,
		Method:  message.SMMethodPatchConfigs,
	})

	if err = m.coreProxy.RouteMessage(ctx, ev); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	log.Debug("dispose message completed", zfield.Eid(base.ID),
		zfield.MsgID(msgID), zfield.Elapsed(elapsedTime.Elapsed()))

	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	return entityToBase(entity), errors.Wrap(err, "patch entity configs")
}

// QueryConfigs query entity configs.
func (m *entityManager) QueryConfigs(ctx context.Context, en *Base, propertyIDs []string) (base *Base, err error) {
	log.Info("entity.PatchConfigs",
		zfield.Eid(base.ID), zfield.Type(base.Type),
		zfield.Owner(base.Owner), zfield.Source(base.Source), zfield.Base(en.JSON()))

	// get entity config file.
	var entity *dao.Entity
	if entity, err = m.entityRepo.GetEntity(ctx, &dao.Entity{ID: base.ID}); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	// get properties by ids.
	// TODO: 实现对ConfigFile的patch.copy操作.

	return entityToBase(entity), nil
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
