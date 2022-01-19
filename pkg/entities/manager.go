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
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/dao"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/core/pkg/tql"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const EntityStateName = "core-state"

type entityManager struct {
	entityRepo   dao.IDao
	daprClient   dapr.Client
	etcdClient   *clientv3.Client
	searchClient pb.SearchHTTPServer
	stateManager statem.StateManager

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(ctx context.Context, stateManager statem.StateManager, searchClient pb.SearchHTTPServer) (EntityManager, error) {
	var (
		err        error
		daprClient dapr.Client
		etcdClient *clientv3.Client
	)

	if daprClient, err = dapr.NewClient(); nil != err {
		return nil, errors.Wrap(err, "create manager failed")
	} else if etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Get().Etcd.Address,
		DialTimeout: 3 * time.Second,
	}); nil != err {
		return nil, errors.Wrap(err, "create manager failed")
	}

	ctx, cancel := context.WithCancel(ctx)

	return &entityManager{
		ctx:          ctx,
		cancel:       cancel,
		daprClient:   daprClient,
		etcdClient:   etcdClient,
		searchClient: searchClient,
		stateManager: stateManager,
		lock:         sync.RWMutex{},
	}, nil
}

func (m *entityManager) Start() error {
	return errors.Wrap(m.stateManager.Start(), "start entity manager")
}

func (m *entityManager) OnMessage(ctx context.Context, msgCtx statem.MessageContext) {
	// 接受来自 pubsub 的消息，这些消息将触发 实体 运行时.
	m.stateManager.HandleMessage(ctx, msgCtx)
}

// ------------------------------------APIs-----------------------------.

// CreateEntity create a entity.
func (m *entityManager) CreateEntity(ctx context.Context, base *Base) (*Base, error) {
	var err error
	if base.ID == "" {
		base.ID = uuid()
	}

	// 1. check entity exists.
	if err = m.entityRepo.Exists(ctx, base.ID); nil != err {
		log.Error("check entity", zap.Error(err), zfield.Eid(base.ID))
		return nil, errors.Wrap(err, "create entity")
	}

	// 2. check template entity.
	templateID, _ := ctx.Value(TemplateEntityID{}).(string)
	if templateID != "" {
		if err = m.entityRepo.Exists(ctx, templateID); nil != err {
			log.Error("check template", zap.Error(err), zfield.Eid(templateID))
			return nil, errors.Wrap(err, "create entity")
		}
	}

	// 3. 向实体发送消息，来在某一个节点上拉起实体，执行实体运行时过程.
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.PropertyMessage{
			StateID:    base.ID,
			Operator:   "replace",
			Properties: base.Properties,
		},
	}

	msgCtx.Headers.SetType(base.Type)
	msgCtx.Headers.SetOwner(base.Owner)
	msgCtx.Headers.SetReceiver(base.ID)

	return base, errors.Wrap(m.stateManager.RouteMessage(ctx, msgCtx), "create entity")
}

// DeleteEntity delete an entity from manager.
func (m *entityManager) DeleteEntity(ctx context.Context, en *Base) (base *Base, err error) {
	// 1. delete from elasticsearch.
	if _, err = m.searchClient.DeleteByID(ctx, &pb.DeleteByIDRequest{Id: en.ID}); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(en.ID))
		if elastic.IsNotFound(err) {
			return nil, errors.Wrap(err, "delete entity from es state")
		}
	}

	// // 2. delete from runtime.
	// if base, err = m.stateManager.DeleteStateMarchin(ctx, en); nil != err {
	// 	log.Error("delete entity runtime", zap.Error(err), zfield.Eid(en.ID))
	// 	if base, err = m.getEntityFromState(ctx, en); nil != err {
	// 		log.Error("get entity", zap.Error(err), zfield.Eid(en.ID))
	// 		return nil, errors.Wrap(err, "delete entity from runtime")
	// 	}
	// }

	// 3. delete from state.
	if err = m.daprClient.DeleteState(ctx, EntityStateName, en.ID); nil != err {
		log.Error("delete entity", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "delete entity from state")
	}

	// 4. delete tql from etcd.
	for _, mm := range base.Mappers {
		if _, err = m.etcdClient.Delete(ctx, util.FormatMapper(base.Type, base.ID, mm.Name), clientv3.WithPrefix()); nil != err {
			log.Error("delete entity mapper", zap.Error(err), zfield.Eid(en.ID), zap.Any("mapper", base.Mappers))
		}
	}

	// 5. log record.
	log.Info("delete entity", zfield.Eid(en.ID), zap.Any("entity", base))

	return base, errors.Wrap(err, "delete entity")
}

// GetProperties returns Base.
func (m *entityManager) GetProperties(ctx context.Context, en *Base) (base *Base, err error) {
	if base, err = m.getEntityFromState(ctx, en); nil != err {
		log.Error("get entity", zap.Error(err), zfield.Eid(en.ID))
	}
	return base, errors.Wrap(err, "entity GetProperties")
}

func (m *entityManager) getEntityFromState(ctx context.Context, en *Base) (base *Base, err error) {
	var item *dapr.StateItem
	if item, err = m.daprClient.GetState(ctx, EntityStateName, en.ID); nil != err {
		return
	} else if nil == item || len(item.Value) == 0 {
		return nil, ErrEntityNotFound
	}
	return
}

// SetProperties set properties into entity.
func (m *entityManager) SetProperties(ctx context.Context, en *Base) (base *Base, err error) {
	// // TODO：这里的调用其实是可能通过entity-manager的proxy的同步调用，这个可以设置可选项.
	// if err = m.stateManager.SetProperties(ctx, en); nil != err {
	// 	log.Error("set entity properties", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "set entity properties")
	// }

	base, err = m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "set entity properties")
}

func (m *entityManager) PatchEntity(ctx context.Context, en *Base, patchData []*pb.PatchData) (base *Base, err error) {
	// if err = m.stateManager.PatchEntity(ctx, en, patchData); nil != err {
	// 	log.Error("patch entity", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "patch entity properties")
	// }

	base, err = m.getEntityFromState(ctx, en)
	for _, pd := range patchData {
		if pd.Operator == constraint.PatchOpCopy.String() {
			if base.Properties[pd.Path], err = base.GetProperty(pd.Path); nil != err {
				log.Error("patch copy config", zap.String("path", pd.Path), zap.String("op", pd.Operator))
			}
		}
	}
	return base, errors.Wrap(err, "patch entity properties")
}

// AppendMapper append a mapper into entity.
func (m *entityManager) AppendMapper(ctx context.Context, en *Base) (base *Base, err error) {
	// 1. 判断实体是否存在.
	if _, err = m.daprClient.GetState(ctx, EntityStateName, en.ID); nil != err {
		log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "get state")
	}

	// check TQLs.
	if err = checkTQLs(en); nil != err {
		return nil, errors.Wrap(err, "check subscription")
	}

	// 2. 将 mapper 推到 etcd.
	for _, mm := range en.Mappers {
		if _, err = m.etcdClient.Put(ctx, util.FormatMapper(en.Type, en.ID, mm.Name), mm.TQLString); nil != err {
			log.Error("append mapper", zap.Error(err), zfield.Eid(en.ID), zap.Any("mapper", mm))
			return nil, errors.Wrap(err, "append mapper")
		}
		log.Info("append mapper", zfield.Eid(en.ID), zap.Any("mapper", mm))
	}

	base, err = m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "append mapper")
}

// DeleteMapper delete mapper from entity.
func (m *entityManager) RemoveMapper(ctx context.Context, en *Base) (base *Base, err error) {
	// 1. 判断实体是否存在.
	if _, err = m.daprClient.GetState(ctx, EntityStateName, en.ID); nil != err {
		log.Error("remove mapper", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "remove mapper")
	}

	// 2. 将 mapper 推到 etcd.
	for _, mm := range en.Mappers {
		if _, err = m.etcdClient.Delete(ctx, util.FormatMapper(en.Type, en.ID, mm.Name)); nil != err {
			log.Error("remove mapper", zap.Error(err), zfield.Eid(en.ID), zap.Any("mapper", mm))
			return nil, errors.Wrap(err, "remove mapper")
		}
		log.Info("remove mapper", zfield.Eid(en.ID), zap.Any("mapper", mm))
	}

	base, err = m.getEntityFromState(ctx, en)
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
	// if err = m.stateManager.SetConfigs(ctx, en); nil != err {
	// 	log.Error("set entity configs", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "set entity configs")
	// }

	base, err = m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "set entity configs")
}

// PatchConfigs patch properties into entity.
func (m *entityManager) PatchConfigs(ctx context.Context, en *Base, patchData []*statem.PatchData) (base *Base, err error) {
	// if err = m.stateManager.PatchConfigs(ctx, en, patchData); nil != err {
	// 	log.Error("patch entity configs", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "patch entity configs")
	// }

	if base, err = m.getEntityFromState(ctx, en); nil != err {
		log.Error("patch entity configs", zap.Error(err), zfield.Eid(en.ID))
		return nil, errors.Wrap(err, "patch entity configs")
	}

	for _, pd := range patchData {
		if pd.Operator == constraint.PatchOpCopy {
			cfg, err0 := base.GetConfig(pd.Path)
			if nil != err0 {
				log.Error("patch copy config", zap.String("path", pd.Path), zap.String("op", pd.Operator.String()))
				continue
			}
			base.Configs[pd.Path] = cfg
		}
	}
	return base, nil
}

// AppendConfigs append entity configs.
func (m *entityManager) AppendConfigs(ctx context.Context, en *Base) (base *Base, err error) {
	// if err = m.stateManager.AppendConfigs(ctx, en); nil != err {
	// 	log.Error("append entity configs", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "append entity configs")
	// }

	base, err = m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "append entity configs")
}

// RemoveConfigs remove entity configs.
func (m *entityManager) RemoveConfigs(ctx context.Context, en *Base, propertyIDs []string) (base *Base, err error) {
	// if err = m.stateManager.RemoveConfigs(ctx, en, propertyIDs); nil != err {
	// 	log.Error("remove entity configs", zap.Error(err), zfield.Eid(en.ID))
	// 	return nil, errors.Wrap(err, "remove entity configs")
	// }

	base, err = m.getEntityFromState(ctx, en)
	return base, errors.Wrap(err, "remove entity configs")
}

// QueryConfigs query entity configs.
func (m *entityManager) QueryConfigs(ctx context.Context, en *Base, propertyIDs []string) (base *Base, err error) {
	base, err = m.getEntityFromState(ctx, en)
	baseEntity := base.Basic()
	for _, propertyID := range propertyIDs {
		cfg, err0 := base.GetConfig(propertyID)
		if nil != err0 {
			log.Error("query configs", zap.Error(err0), zfield.Eid(en.ID), zap.String("property", propertyID))
			continue
		}
		baseEntity.Configs[propertyID] = cfg
	}
	return &baseEntity, errors.Wrap(err, "remove entity configs")
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
		if tqlInst, err = tql.NewTQL(mm.TQLString); nil != err {
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

// uuid generate an uuid.
func uuid() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// see section 4.1.1.
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3.
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
