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
	"encoding/json"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfiled "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type Manager struct {
	containers    map[string]*Container
	msgCh         chan statem.MessageContext
	disposeCh     chan statem.MessageContext
	coroutinePool *ants.Pool
	actorEnv      environment.IEnvironment

	repository    dao.IDao
	daprClient    dapr.Client
	etcdClient    *clientv3.Client
	searchClient  pb.SearchHTTPServer
	tseriesClient tseries.TimeSerier

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, coroutinePool *ants.Pool, searchClient pb.SearchHTTPServer) (statem.StateManager, error) {
	var (
		daprClient dapr.Client
		etcdClient *clientv3.Client
		err        error
	)

	expireTime := 3 * time.Second
	etcdAddr := config.Get().Etcd.Address
	tseriesClient := tseries.NewTimeSerier(config.Get().TimeSeries.Name)
	returnErr := func(err error) error { return errors.Wrap(err, "new runtime.Manager") }

	if daprClient, err = dapr.NewClient(); nil != err {
		log.Error("")
		return nil, returnErr(err)
	}
	if err = tseriesClient.Init(resource.ParseFrom(config.Get().TimeSeries)); nil != err {
		return nil, returnErr(err)
	}
	if etcdClient, err = clientv3.New(clientv3.Config{Endpoints: etcdAddr, DialTimeout: expireTime}); nil != err {
		return nil, returnErr(err)
	}

	ctx, cancel := context.WithCancel(ctx)

	stateManager := &Manager{
		ctx:           ctx,
		cancel:        cancel,
		daprClient:    daprClient,
		etcdClient:    etcdClient,
		searchClient:  searchClient,
		tseriesClient: tseriesClient,
		actorEnv:      environment.NewEnvironment(),
		containers:    make(map[string]*Container),
		msgCh:         make(chan statem.MessageContext, 10),
		disposeCh:     make(chan statem.MessageContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}

	// set default container.
	stateManager.containers["default"] = NewContainer()
	return stateManager, nil
}

func (m *Manager) init() error {
	// load all subscriptions.
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	log.Info("initialize actor manager, tql loadding...")
	res, err := m.etcdClient.Get(ctx, util.EtcdMapperPrefix, clientv3.WithPrefix())
	if nil != err {
		return errors.Wrap(err, "load all tql")
	}

	pairs := make([]environment.EtcdPair, len(res.Kvs))
	for index, kv := range res.Kvs {
		pairs[index] = environment.EtcdPair{Key: string(kv.Key), Value: kv.Value}
	}

	for _, info := range m.actorEnv.StoreMappers(pairs) {
		log.Debug("load state machine", zfiled.ID(info.EntityID), zap.String("type", info.Type))
		if err = m.loadActor(context.Background(), info.Type, info.EntityID); nil != err {
			log.Error("load state machine", zap.Error(err),
				zap.String("type", info.Type), zfiled.ID(info.EntityID))
		}
	}

	return nil
}

func (m *Manager) watchResource() error {
	// watch tqls.
	tqlWatcher, err := util.NewWatcher(m.ctx, config.Get().Etcd.Address)
	if nil != err {
		return errors.Wrap(err, "create tql watcher failed")
	}

	tqlWatcher.Watch(util.EtcdMapperPrefix, true, func(ev *clientv3.Event) {
		pair := environment.EtcdPair{Key: string(ev.Kv.Key), Value: ev.Kv.Value}
		effects, _ := m.actorEnv.OnMapperChanged(ev.Type, pair)
		m.reloadActor(effects)
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
				eid := msgCtx.Headers.GetReceiver()
				channelID := msgCtx.Headers.Get(statem.MsgCtxHeaderChannelID)
				log.Debug("dispose message", zfiled.ID(eid), zfiled.Message(msgCtx))
				channelID, stateMachine := m.getStateMachine(channelID, eid)
				if nil == stateMachine {
					var err error
					en := &dao.Entity{
						ID:     eid,
						Owner:  msgCtx.Headers.GetOwner(),
						Source: msgCtx.Headers.GetSource(),
						Type:   msgCtx.Headers.Get(statem.MsgCtxHeaderType),
					}
					stateMachine, err = m.loadOrCreate(m.ctx, channelID, true, en)
					if nil != err {
						log.Error("dispatching message", zap.Error(err),
							zfiled.ID(eid), zap.String("channel", channelID), zfiled.Message(msgCtx))
						continue
					}
				}

				if stateMachine.OnMessage(msgCtx.Message) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(stateMachine.HandleLoop)
				}
			case <-m.shutdown:
				log.Info("state machine manager exit.")
				return
			}
		}
	}()

	return nil
}

func (m *Manager) Shutdown() error {
	m.cancel()
	m.shutdown <- struct{}{}
	return nil
}

func (m *Manager) RouteMessage(ctx context.Context, msgCtx statem.MessageContext) error {
	// assume single node.
	return m.HandleMessage(ctx, msgCtx)
}

func (m *Manager) HandleMessage(ctx context.Context, msgCtx statem.MessageContext) error {
	bytes, _ := json.Marshal(msgCtx)
	log.Debug("actor send message", zap.String("msg", string(bytes)))

	// 解耦actor之间的直接调用
	m.msgCh <- msgCtx
	return nil
}

// Resource return resource manager.
func (m *Manager) Resource() statem.ResourceManager {
	panic("implement me")
}

func (m *Manager) getStateMachine(cid, eid string) (string, statem.StateMachiner) {
	if cid == "" {
		cid = "default"
	}

	if container, ok := m.containers[cid]; ok {
		if sm := container.Get(eid); nil != sm {
			if sm.GetStatus() == statem.SMStatusDeleted {
				container.Remove(eid)
				return cid, nil
			}
			return cid, sm
		}
	}

	for channelID, container := range m.containers {
		if sm := container.Get(eid); sm != nil {
			if sm.GetStatus() == statem.SMStatusDeleted {
				container.Remove(eid)
				return cid, nil
			}

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

func (m *Manager) isThisNode() bool {
	return true
}

func (m *Manager) reloadActor(stateIDs []string) error {
	// 判断 actor 是否在当前节点.
	if m.isThisNode() {
		var err error
		for _, stateID := range stateIDs {
			var stateMachine statem.StateMachiner
			base := &dao.Entity{ID: stateID, Type: SMTypeBasic}
			if _, stateMachine = m.getStateMachine("", stateID); nil != stateMachine {
				log.Warn("load state machine", zfiled.ID(stateID))
			} else if stateMachine, err = m.loadOrCreate(m.ctx, "", false, base); nil == err {
				continue
			}
			actorEnv := m.actorEnv.GetActorEnv(stateID)
			stateMachine.WithContext(statem.NewContext(stateMachine, actorEnv.Mappers, actorEnv.Tentacles))
		}
	}
	return nil
}

func (m *Manager) loadActor(ctx context.Context, typ string, id string) error {
	_, err := m.loadOrCreate(ctx, "", false, &dao.Entity{ID: id, Type: typ})
	return errors.Wrap(err, "load entity")
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, flagCreate bool, base *dao.Entity) (sm statem.StateMachiner, err error) {
	log.Debug("load or create actor", zfiled.ID(base.ID),
		zap.String("type", base.Type), zap.String("owner", base.Owner), zap.String("source", base.Source))

	var en *dao.Entity
	if en, err = m.repository.Get(ctx, base.ID); nil != err {
		base = en
	} else {
		log.Warn("load or create actor", zap.Error(err),
			zfiled.Eid(base.ID), zfiled.Type(base.Type), zfiled.Template(base.TemplateID))

		// notfound.
		if !flagCreate || !errors.Is(err, xerrors.ErrEntityNotFound) {
			return nil, errors.Wrap(err, "load or create actor")
		}
	}

	switch base.Type {
	case SMTypeSubscription:
		if sm, err = subscription.NewSubscription(ctx, m, base); nil != err {
			return nil, errors.Wrap(err, "load subscription")
		}
	default:
		// default base entity type.
		if sm, err = statem.NewState(ctx, m, base, nil); nil != err {
			return nil, errors.Wrap(err, "load state machine")
		}
	}

	if channelID == "" {
		channelID = "defult"
	}

	if _, has := m.containers[channelID]; !has {
		m.containers[channelID] = NewContainer()
	}

	thisActorEnv := m.actorEnv.GetActorEnv(sm.GetID())
	sm.WithContext(statem.NewContext(sm, thisActorEnv.Mappers, thisActorEnv.Tentacles))

	m.containers[channelID].Add(sm)
	return sm, nil
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return []string{expression}
}
