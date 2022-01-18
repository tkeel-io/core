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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	zfiled "github.com/tkeel-io/core/pkg/logger"
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
				eid := msgCtx.Headers.GetTargetID()
				channelID := msgCtx.Headers.Get(statem.MessageCtxHeaderChannelID)
				log.Debug("dispose message", zfiled.ID(eid), zfiled.Message(msgCtx))
				channelID, stateMachine := m.getStateMachine(channelID, eid)
				if nil == stateMachine {
					var err error
					en := &statem.Base{
						ID:     eid,
						Owner:  msgCtx.Headers.GetOwner(),
						Source: msgCtx.Headers.GetSource(),
						Type:   msgCtx.Headers.Get(statem.MessageCtxHeaderType),
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

// GetResource return resource manager.
func (m *Manager) GetResource() statem.ResourceManager {
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
			base := &statem.Base{ID: stateID, Type: StateMachineTypeBasic}
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
	_, err := m.loadOrCreate(ctx, "", false, &statem.Base{ID: id, Type: typ})
	return errors.Wrap(err, "load entity")
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, flagCreate bool, base *statem.Base) (sm statem.StateMachiner, err error) {
	var en *statem.Base
	var res *dapr.StateItem
	res, err = m.daprClient.GetState(ctx, EntityStateName, base.ID)

	if nil != err && !flagCreate {
		return nil, errors.Wrap(err, "load state machine")
	} else if en, err = statem.DecodeBase(res.Value); nil == err {
		base = en // decode value to statem.Base.
	} else if !flagCreate {
		return nil, errors.Wrap(err, "load state machine, state not found")
	}

	log.Debug("load or create state machiner",
		zfiled.ID(base.ID),
		zap.String("type", base.Type),
		zap.String("owner", base.Owner),
		zap.String("source", base.Source))

	switch base.Type {
	case StateMachineTypeSubscription:
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
			Properties: en.Properties,
		},
	}
	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)
	msgCtx.Headers.SetSource(en.Source)
	msgCtx.Headers.Set(statem.MessageCtxHeaderType, en.Type)

	m.HandleMessage(ctx, msgCtx)

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
			m.HandleMessage(ctx, msgCtx)
		}
	}

	return nil
}

// SetConfigs set entity configs.
func (m *Manager) SetConfigs(ctx context.Context, en *statem.Base) error {
	var (
		err          error
		channelID    string
		stateMachine statem.StateMachiner
	)

	// load state machine.
	if channelID, stateMachine = m.getStateMachine("", en.ID); nil == stateMachine {
		if stateMachine, err = m.loadOrCreate(ctx, channelID, false, en); nil != err {
			log.Error("set configs",
				zfiled.ID(en.ID),
				zap.Any("entity", en),
				zap.String("channel", channelID))
			return errors.Wrap(err, "set entity configs")
		}
	}

	// flush entity configs.
	return errors.Wrap(stateMachine.Flush(ctx), "set entity configs")
}

// PatchConfigs patch entity configs.
func (m *Manager) PatchConfigs(ctx context.Context, en *statem.Base, patchData []*statem.PatchData) error {
	var (
		err          error
		channelID    string
		stateMachine statem.StateMachiner
	)

	// load state machine.
	if channelID, stateMachine = m.getStateMachine("", en.ID); nil == stateMachine {
		if stateMachine, err = m.loadOrCreate(ctx, channelID, false, en); nil != err {
			log.Error("set configs",
				zfiled.ID(en.ID),
				zap.Any("entity", en),
				zap.String("channel", channelID))
			return errors.Wrap(err, "set entity configs")
		}
	}

	// flush entity configs.
	return errors.Wrap(stateMachine.Flush(ctx), "set entity configs")
}

// AppendConfigs append entity configs.
func (m *Manager) AppendConfigs(ctx context.Context, en *statem.Base) error {
	var (
		err          error
		channelID    string
		stateMachine statem.StateMachiner
	)

	// load state machine.
	if channelID, stateMachine = m.getStateMachine("", en.ID); nil == stateMachine {
		if stateMachine, err = m.loadOrCreate(ctx, channelID, false, en); nil != err {
			log.Error("append configs",
				zfiled.ID(en.ID),
				zap.Any("entity", en),
				zap.String("channel", channelID))
			return errors.Wrap(err, "append entity configs")
		}
	}

	// flush entity configs.
	return errors.Wrap(stateMachine.Flush(ctx), "append entity configs")
}

// RemoveConfigs remove entity configs.
func (m *Manager) RemoveConfigs(ctx context.Context, en *statem.Base, propertyIDs []string) error {
	var (
		err          error
		channelID    string
		stateMachine statem.StateMachiner
	)

	// load state machine.
	if channelID, stateMachine = m.getStateMachine("", en.ID); nil == stateMachine {
		if stateMachine, err = m.loadOrCreate(ctx, channelID, false, en); nil != err {
			log.Error("remove configs",
				zfiled.ID(en.ID),
				zap.Any("entity", en),
				zap.String("channel", channelID))
			return errors.Wrap(err, "remove entity configs")
		}
	}

	// flush entity configs.
	return errors.Wrap(stateMachine.Flush(ctx), "remove entity configs")
}

// DeleteStateMachine delete runtime.Entity.
func (m *Manager) DeleteStateMarchin(ctx context.Context, base *statem.Base) (*statem.Base, error) {
	var err error
	channelID, stateMachine := m.getStateMachine("", base.ID)
	if nil == stateMachine {
		if stateMachine, err = m.loadOrCreate(m.ctx, channelID, true, base); nil != err {
			log.Error("remove configs",
				zfiled.ID(base.ID),
				zap.Any("entity", base),
				zap.String("channel", channelID))
			return nil, errors.Wrap(err, "remove entity configs")
		}
	}

	return stateMachine.GetBase(), nil
}

// CleanEntity clean entity.
func (m *Manager) CleanEntity(ctx context.Context, id string) error {
	channelID, sm := m.getStateMachine("", id)
	if nil != sm {
		m.containers[channelID].Remove(id)
	}
	return nil
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
