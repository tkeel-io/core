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

package runtime

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type Manager struct {
	containers    map[string]*Container
	msgCh         chan statem.MessageContext
	disposeCh     chan statem.MessageContext
	coroutinePool *ants.Pool
	actorEnv      *Environment

	daprClient   dapr.Client
	etcdClient   *clientv3.Client
	searchClient pb.SearchHTTPServer

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, coroutinePool *ants.Pool, searchClient pb.SearchHTTPServer) (*Manager, error) {
	var (
		err        error
		daprClient dapr.Client
		etcdClient *clientv3.Client
	)

	if daprClient, err = dapr.NewClient(); nil != err {
		return nil, errors.Wrap(err, "create manager failed")
	} else if etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   config.GetConfig().Etcd.Address,
		DialTimeout: 3 * time.Second,
	}); nil != err {
		return nil, errors.Wrap(err, "create manager failed")
	}

	ctx, cancel := context.WithCancel(ctx)

	mgr := &Manager{
		ctx:           ctx,
		cancel:        cancel,
		actorEnv:      NewEnv(),
		daprClient:    daprClient,
		etcdClient:    etcdClient,
		searchClient:  searchClient,
		containers:    make(map[string]*Container),
		msgCh:         make(chan statem.MessageContext, 10),
		disposeCh:     make(chan statem.MessageContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}

	// set default container.
	mgr.containers["default"] = NewContainer()
	return mgr, nil
}

func (m *Manager) SendMsg(msgCtx statem.MessageContext) {
	// 解耦actor之间的直接调用
	m.msgCh <- msgCtx
}

func (m *Manager) init() error {
	// load all subcriptions.
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	log.Info("initialize actor manager, tql loadding...")
	res, err := m.etcdClient.Get(ctx, TQLEtcdPrefix, clientv3.WithPrefix())
	if nil != err {
		return errors.Wrap(err, "load all tql")
	}

	descs := make([]EtcdPair, len(res.Kvs))
	for index, kv := range res.Kvs {
		descs[index] = EtcdPair{Key: string(kv.Key), Value: kv.Value}
		log.Info("load tql", zap.String("key", string(kv.Key)), zap.String("tql", string(kv.Value)))
	}

	loadEntities := m.actorEnv.LoadMapper(descs)
	for _, info := range loadEntities {
		log.Info("load entity", logger.EntityID(info.EntityID), zap.String("type", info.Type))
		if err = m.loadActor(context.Background(), info.Type, info.EntityID); nil != err {
			log.Error("load entity", zap.Error(err), logger.EntityID(info.EntityID), zap.String("type", info.Type))
		}
	}

	return nil
}

func (m *Manager) watchResource() error {
	// watch tqls.
	tqlWatcher, err := util.NewWatcher(m.ctx, config.GetConfig().Etcd.Address)
	if nil != err {
		return errors.Wrap(err, "create tql watcher failed")
	}

	tqlWatcher.Watch(TQLEtcdPrefix, true, func(ev *clientv3.Event) {
		// on changed.
		m.actorEnv.OnMapperChanged(ev.Type, EtcdPair{Key: string(ev.Kv.Key), Value: ev.Kv.Value})
	})

	return nil
}

func (m *Manager) Start() error {
	// init: load some resource.
	m.init()
	// watch resource.
	m.watchResource()

	go func() {
		for {
			select {
			case <-m.ctx.Done():
				log.Info("entity manager exited.")
				return
			case msgCtx := <-m.msgCh:
				// dispatch message. 将消息分发到不同的节点。
				m.disposeCh <- msgCtx

			case msgCtx := <-m.disposeCh:
				eid := msgCtx.Headers.GetTargetID()
				channelID := msgCtx.Headers.Get(statem.MessageCtxHeaderChannelID)
				log.Info("dispose message", logger.EntityID(eid), logger.MessageInst(msgCtx))
				channelID, stateMarchine := m.getStateMarchine(channelID, eid)
				if nil == stateMarchine {
					var err error
					en := &statem.Base{
						ID:     eid,
						Owner:  msgCtx.Headers.GetOwner(),
						Source: msgCtx.Headers.GetSource(),
						Type:   msgCtx.Headers.Get(statem.MessageCtxHeaderType),
					}
					stateMarchine, err = m.loadOrCreate(m.ctx, channelID, en)
					if nil != err {
						log.Error("dispatching message", zap.Error(err),
							logger.EntityID(eid), zap.String("channel", channelID), logger.MessageInst(msgCtx))
						continue
					}
				}

				if stateMarchine.OnMessage(msgCtx.Message) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(stateMarchine.HandleLoop)
				}
			case <-m.shutdown:
				log.Info("state marchine manager exit.")
				return
			}
		}
	}()

	return nil
}

func (m *Manager) Shutdown() {
	m.cancel()
	m.shutdown <- struct{}{}
}

func (m *Manager) GetDaprClient() dapr.Client {
	return m.daprClient
}

func (m *Manager) getStateMarchine(cid, eid string) (string, statem.StateMarchiner) {
	if cid == "" {
		cid = "default"
	}

	if container, ok := m.containers[cid]; ok {
		if sm := container.Get(eid); nil != sm {
			return cid, sm
		}
	}

	for channelID, container := range m.containers {
		if sm := container.Get(eid); sm != nil {
			if channelID == "default" && cid != channelID {
				container.Remove(sm.GetID())
				if _, ok := m.containers[cid]; !ok {
					m.containers[cid] = NewContainer()
				}
				m.containers[cid].Add(sm)
			}
			return cid, sm
		}
	}

	return cid, nil
}

func (m *Manager) loadActor(ctx context.Context, typ string, id string) error {
	_, err := m.loadOrCreate(ctx, "", &statem.Base{
		ID:   id,
		Type: typ,
	})
	return errors.Wrap(err, "load entity")
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, base *statem.Base) (sm statem.StateMarchiner, err error) {
	var res *dapr.StateItem
	switch base.Type {
	case StateMarchineTypeSubscription:
		if res, err = m.daprClient.GetState(ctx, EntityStateName, base.ID); nil != err {
			// TODO: 订阅不存在，所以应该通知被订阅方取消订阅.
			return nil, errors.Wrap(err, "load subscription")
		} else if base, err = statem.DecodeBase(res.Value); nil != err {
			return nil, errors.Wrap(err, "load subscription")
		}
		sm, err = newSubscription(ctx, m, base)
	default:
		// default base entity type.
		if res, err = m.daprClient.GetState(ctx, EntityStateName, base.ID); nil != err {
			log.Warn("load state", zap.Error(err), logger.EntityID(base.ID))
		} else if en, errr := statem.DecodeBase(res.Value); nil == errr {
			base = en
		} else {
			log.Error("load or create state",
				zap.String("channel", channelID), logger.EntityID(base.ID), zap.Error(err))
		}
		sm, err = statem.NewState(ctx, m, base, nil)
	}

	if nil != err {
		return nil, errors.Wrap(err, "create state runtime")
	}

	if channelID == "" {
		channelID = "defult"
	}

	if _, has := m.containers[channelID]; !has {
		m.containers[channelID] = NewContainer()
	}

	sm.Setup()
	m.containers[channelID].Add(sm)
	return sm, nil
}

func (m *Manager) HandleMsg(ctx context.Context, msg statem.MessageContext) {
	// dispose message from pubsub.
	m.msgCh <- msg
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return []string{expression}
}

// ------------------------------------APIs-----------------------------.

// SetProperties set properties into entity.
func (m *Manager) SetProperties(ctx context.Context, en *statem.Base) error {
	if en.ID == "" {
		en.ID = uuid()
	}

	// set properties.
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.PropertyMessage{
			StateID:    en.ID,
			Operator:   constraint.PatchOpReplace.String(),
			Properties: en.KValues,
		},
	}
	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)
	msgCtx.Headers.SetSource(en.Source)
	msgCtx.Headers.Set(statem.MessageCtxHeaderType, en.Type)

	m.SendMsg(msgCtx)

	return nil
}

func (m *Manager) PatchEntity(ctx context.Context, en *statem.Base, patchData []*pb.PatchData) error {
	pdm := make(map[string][]*pb.PatchData)
	for _, pd := range patchData {
		pdm[pd.Operator] = append(pdm[pd.Operator], pd)
	}

	for op, pds := range pdm {
		kvs := make(map[string]constraint.Node)
		for _, pd := range pds {
			kvs[pd.Path] = constraint.NewNode(pd.Value.AsInterface())
		}

		if len(kvs) > 0 {
			msgCtx := statem.MessageContext{
				Headers: statem.Header{},
				Message: statem.PropertyMessage{
					StateID:    en.ID,
					Operator:   op,
					Properties: kvs,
				},
			}

			// set headers.
			msgCtx.Headers.SetOwner(en.Owner)
			msgCtx.Headers.SetTargetID(en.ID)
			msgCtx.Headers.Set(statem.MessageCtxHeaderType, en.Type)
			m.SendMsg(msgCtx)
		}
	}

	return nil
}

// SetProperties set properties into entity.
func (m *Manager) SetConfigs(ctx context.Context, en *statem.Base) error {
	if en.ID == "" {
		en.ID = uuid()
	}

	channelID, stateMarchine := m.getStateMarchine("", en.ID)
	if nil == stateMarchine {
		var err error
		if stateMarchine, err = m.loadOrCreate(m.ctx, channelID, en); nil != err {
			log.Error("dispatching message", logger.EntityID(en.ID), zap.String("channel", channelID), zap.Any("entity", en))
			return errors.Wrap(err, "runtime.setconfigs")
		}
	}

	err := stateMarchine.SetConfig(en.Configs)
	return errors.Wrap(err, "runtime.setconfigs")
}

func (m *Manager) DeleteStateMarchin(ctx context.Context, base *statem.Base) (*statem.Base, error) {
	sm, err := m.loadOrCreate(ctx, "", base)
	if nil != err {
		return nil, errors.Wrap(err, "runtime.delete state marchine")
	}
	sm.SetStatus(statem.SMStatusDeleted)
	return sm.GetBase(), nil
}

func (m *Manager) CleanEntity(ctx context.Context, id string) error {
	channelID, sm := m.getStateMarchine("", id)
	if nil != sm {
		m.containers[channelID].Remove(id)
	}
	return nil
}

// AppendMapper append a mapper into entity.
func (m *Manager) AppendMapper(ctx context.Context, en *statem.Base) error {
	if len(en.Mappers) == 0 {
		log.Error("append mapper into entity failed.", logger.EntityID(en.ID), zap.Error(ErrInvalidParams))
		return errors.Wrap(ErrInvalidParams, "append entity mapper failed")
	}

	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.MapperMessage{
			Operator: statem.MapperOperatorAppend,
			Mapper:   en.Mappers[0],
		},
	}

	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)

	m.SendMsg(msgCtx)

	return nil
}

// DeleteMapper delete mapper from entity.
func (m *Manager) RemoveMapper(ctx context.Context, en *statem.Base) error {
	if len(en.Mappers) == 0 {
		log.Error("remove mapper failed.", logger.EntityID(en.ID), zap.Error(ErrInvalidParams))
		return errors.Wrap(ErrInvalidParams, "remove entity mapper failed")
	}

	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.MapperMessage{
			Operator: statem.MapperOperatorRemove,
			Mapper:   en.Mappers[0],
		},
	}

	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)

	m.SendMsg(msgCtx)

	return nil
}

func (m *Manager) SearchFlush(ctx context.Context, values map[string]interface{}) error {
	var err error
	var val *structpb.Value
	if val, err = structpb.NewValue(values); nil != err {
		log.Error("search index failed.", zap.Error(err))
	} else if _, err = m.searchClient.Index(ctx, &pb.IndexObject{Obj: val}); nil != err {
		log.Error("search index failed.", zap.Error(err))
	}
	return errors.Wrap(err, "SearchFlushfailed")
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
