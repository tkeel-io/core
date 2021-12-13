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

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type Manager struct {
	containers    map[string]*Container
	msgCh         chan statem.MessageContext
	disposeCh     chan statem.MessageContext
	coroutinePool *ants.Pool

	daprClient   dapr.Client
	searchClient pb.SearchHTTPServer

	shutdown chan struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context, coroutinePool *ants.Pool, searchClient pb.SearchHTTPServer) (*Manager, error) {
	daprClient, err := dapr.NewClient()
	if nil != err {
		return nil, errors.Wrap(err, "create manager failed")
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Manager{
		ctx:           ctx,
		cancel:        cancel,
		daprClient:    daprClient,
		searchClient:  searchClient,
		containers:    make(map[string]*Container),
		msgCh:         make(chan statem.MessageContext, 10),
		disposeCh:     make(chan statem.MessageContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}, nil
}

func (m *Manager) SendMsg(msgCtx statem.MessageContext) {
	// 解耦actor之间的直接调用
	m.msgCh <- msgCtx
}

func (m *Manager) Start() error {
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
				log.Info("dispose message",
					logger.EntityID(msgCtx.Headers.GetTargetID()), logger.MessageInst(msgCtx))
				eid := msgCtx.Headers.GetTargetID()
				channelID := msgCtx.Headers.Get(statem.MessageCtxHeaderChannelID)
				channelID, stateMarchine := m.getStateMarchine(channelID, eid)
				if nil == stateMarchine {
					var err error
					en := &statem.Base{
						ID:     msgCtx.Headers.GetTargetID(),
						Owner:  msgCtx.Headers.GetOwner(),
						Source: msgCtx.Headers.GetSource(),
						Type:   msgCtx.Headers.Get(statem.MessageCtxHeaderType),
					}
					stateMarchine, err = m.loadOrCreate(m.ctx, channelID, en)
					if nil != err {
						log.Error("dispatching message",
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

func (m *Manager) getStateMarchine(cid, eid string) (string, statem.StateMarchiner) {
	if cid == "" {
		cid = "default"
	}

	if container, ok := m.containers[cid]; ok {
		return cid, container.states[eid]
	}

	for channelID, container := range m.containers {
		if sm := container.Get(eid); sm != nil {
			return channelID, sm
		}
	}

	return cid, nil
}

func (m *Manager) loadOrCreate(ctx context.Context, channelID string, base *statem.Base) (sm statem.StateMarchiner, err error) {
	// load state-marchine from state store.

	// create from base.
	// 临时创建
	switch base.Type {
	case StateMarchineTypeSubscription:
		// subscription entity type.
		sm, err = newSubscription(ctx, m, base)
	default:
		// default base entity type.
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

	m.containers[channelID].Add(sm)
	return sm, nil
}

func (m *Manager) HandleMsg(ctx context.Context, msg statem.MessageContext) {
	// dispose message from pubsub.
	m.msgCh <- msg
}

// Tools.

func (m *Manager) EscapedEntities(expression string) []string {
	return nil
}

// ------------------------------------APIs-----------------------------.

// SetProperties set properties into entity.
func (m *Manager) SetProperties(ctx context.Context, en *statem.Base) error {
	if en.ID == "" {
		en.ID = uuid()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// set properties.
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.PropertyMessage{
			StateID:    en.ID,
			Operator:   constraint.PatchOpReplace.String(),
			Properties: en.KValues,
			MessageBase: statem.MessageBase{
				PromiseHandler: func(v interface{}) {
					wg.Done()
				},
			},
		},
	}
	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)
	msgCtx.Headers.SetSource(en.Source)
	msgCtx.Headers.Set(statem.MessageCtxHeaderType, en.Type)

	m.SendMsg(msgCtx)

	wg.Wait()

	return nil
}

func (m *Manager) PatchEntity(ctx context.Context, en *statem.Base, patchData []*pb.PatchData) error {
	wg := &sync.WaitGroup{}
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
			wg.Add(1)
			msgCtx := statem.MessageContext{
				Headers: statem.Header{},
				Message: statem.PropertyMessage{
					StateID:    en.ID,
					Operator:   op,
					Properties: kvs,
					MessageBase: statem.MessageBase{
						PromiseHandler: func(v interface{}) {
							wg.Done()
						},
					},
				},
			}

			// set headers.
			msgCtx.Headers.SetOwner(en.Owner)
			msgCtx.Headers.SetTargetID(en.ID)
			msgCtx.Headers.Set(statem.MessageCtxHeaderType, en.Type)
			m.SendMsg(msgCtx)
		}
	}

	wg.Wait()

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

	wg := &sync.WaitGroup{}
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.MapperMessage{
			Operator: statem.MapperOperatorAppend,
			Mapper:   en.Mappers[0],
		},
	}

	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)

	wg.Add(1)
	m.SendMsg(msgCtx)

	wg.Wait()

	return nil
}

// DeleteMapper delete mapper from entity.
func (m *Manager) RemoveMapper(ctx context.Context, en *statem.Base) error {
	if len(en.Mappers) == 0 {
		log.Error("remove mapper failed.", logger.EntityID(en.ID), zap.Error(ErrInvalidParams))
		return errors.Wrap(ErrInvalidParams, "remove entity mapper failed")
	}

	wg := &sync.WaitGroup{}
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.MapperMessage{
			Operator: statem.MapperOperatorRemove,
			Mapper:   en.Mappers[0],
		},
	}

	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)

	wg.Add(1)
	m.SendMsg(msgCtx)

	wg.Wait()
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
