package entities

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/statem"
)

type EntityManager struct {
	entities      map[string]EntityOp
	msgCh         chan statem.MessageContext
	disposeCh     chan statem.MessageContext
	coroutinePool *ants.Pool

	daprClient dapr.Client

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(ctx context.Context, coroutinePool *ants.Pool) (*EntityManager, error) {
	daprClient, err := dapr.NewClient()
	if nil != err {
		return nil, errors.Wrap(err, "create entity manager failed")
	}

	ctx, cancel := context.WithCancel(ctx)

	return &EntityManager{
		ctx:           ctx,
		cancel:        cancel,
		daprClient:    daprClient,
		entities:      make(map[string]EntityOp),
		msgCh:         make(chan statem.MessageContext, 10),
		disposeCh:     make(chan statem.MessageContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}, nil
}

func (m *EntityManager) SendMsg(msgCtx statem.MessageContext) {
	// 解耦actor之间的直接调用
	m.msgCh <- msgCtx
}

func (m *EntityManager) Start() error {
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				log.Info("entity EntityManager exited.")
				return
			case msgCtx := <-m.msgCh:
				// dispatch message. 将消息分发到不同的节点。
				m.disposeCh <- msgCtx

			case msgCtx := <-m.disposeCh:
				// invoke msg.
				// 实际上reactor模式有一个致命的问题就是消息乱序, 引入mailbox可以有效规避乱序问题.
				// 消费的消息统一来自Inbox，不存在无entityID的情况.
				// 如果entity在当前节点不存在就将entity调度到当前节点.
				log.Infof("dispose message failed, entity: %s", msgCtx.Headers.GetTargetID())
				eid := msgCtx.Headers.GetTargetID()
				_, has := m.entities[eid]
				if !has {
					// rebalance entity.
					en := &statem.Base{
						ID:    msgCtx.Headers.GetTargetID(),
						Owner: msgCtx.Headers.GetOwner(),
						Type:  msgCtx.Headers.GetDefault(MessageCtxHeaderEntityType, EntityTypeBaseEntity),
					}

					if err := m.rebalanceEntity(context.Background(), en); nil != err {
						log.Errorf("dispose message failed, err: %s", err.Error())
						continue
					}

					log.Infof("rebalance entity(%s)", eid)
				}

				enInst := m.entities[eid]
				if enInst.OnMessage(msgCtx.Message) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(enInst.HandleLoop)
				}
			}
		}
	}()

	return nil
}

func (m *EntityManager) HandleMsg(ctx context.Context, msg statem.MessageContext) {
	// dispose message from pubsub.
	m.msgCh <- msg
}

func (m *EntityManager) rebalanceEntity(ctx context.Context, en *statem.Base) error {
	// 1. 通过placement查询entity是否在当前节点.
	// 2.1. 如果在当前节点则在当前节点创建该实体.
	// 2.2 如果实体不属于当前节点，则将消息转发出去.

	var (
		err        error
		entityInst EntityOp
	)

	// 从状态存储中获取实体信息.
	// TODO: 这里从状态存储中拿到实体信息.

	// 临时创建
	switch en.Type {
	case EntityTypeSubscription:
		// subscription entity type.
		entityInst, err = newSubscription(context.Background(), m, en)
	default:
		// default base entity type.
		entityInst, err = newEntity(context.Background(), m, en)
	}

	if nil != err {
		return errors.Wrap(err, "rrebalance entity failed")
	}

	m.entities[entityInst.GetID()] = entityInst
	return nil
}

// Tools.

func (m *EntityManager) EscapedEntities(expression string) []string {
	return nil
}

// ------------------------------------APIs-----------------------------.

// DeleteEntity delete an entity from manager.
func (m *EntityManager) DeleteEntity(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if _, has := m.entities[en.ID]; !has {
		log.Errorf("DeleteEntity failed, entity(%s), err: %s", en.ID, errEntityNotFound.Error())
		return nil, errEntityNotFound
	}

	m.entities[en.ID].GetBase().Status = statem.StateStatusDeleted
	enObj := m.entities[en.ID].GetBase().Copy()

	return &enObj, nil
}

// GetProperties returns statem.Base.
func (m *EntityManager) GetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	// just for standalone.
	if _, has := m.entities[en.ID]; !has {
		log.Errorf("GetProperties failed, entity(%s), err: %s", en.ID, errEntityNotFound.Error())
		return nil, errEntityNotFound
	}

	enObj := m.entities[en.ID].GetBase().Copy()

	return &enObj, nil
}

// SetProperties set properties into entity.
func (m *EntityManager) SetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if en.ID == "" {
		en.ID = uuid()
	}

	// set properties.
	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.PropertyMessage{
			StateID:    en.ID,
			Properties: en.KValues,
		},
	}

	msgCtx.Headers.SetOwner(en.Owner)
	msgCtx.Headers.SetTargetID(en.ID)
	msgCtx.Headers.Set(MessageCtxHeaderEntityType, en.Type)

	m.SendMsg(msgCtx)

	return nil, nil
}

// SetProperties set properties into entity.
func (m *EntityManager) SetConfigs(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if en.ID == "" {
		en.ID = uuid()
	}

	// set configs.
	// 如果不存在实体则创建.
	// 如果实体在当前节点则直接调用设置Configs.
	// 如果实体不在当前节点则调用rpc同步.
	_, exists := m.entities[en.ID]
	if !exists {
		// 临时直接创建.
		m.rebalanceEntity(ctx, en)
	}

	m.entities[en.ID].SetConfig(en.Configs)

	if len(en.KValues) > 0 {
		msgCtx := statem.MessageContext{
			Headers: statem.Header{},
			Message: statem.PropertyMessage{
				StateID:    en.ID,
				Properties: en.KValues,
			},
		}

		msgCtx.Headers.SetOwner(en.Owner)
		msgCtx.Headers.SetTargetID(en.ID)
		msgCtx.Headers.Set(MessageCtxHeaderEntityType, en.Type)

		m.SendMsg(msgCtx)
	}

	return nil, nil
}

// AppendMapper append a mapper into entity.
func (m *EntityManager) AppendMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if len(en.Mappers) == 0 {
		log.Errorf("append mapper into entity failed, %s", errEmptyEntityMapper)
		return nil, errors.Wrap(errEmptyEntityMapper, "append entity mapper failed")
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
	return en, nil
}

// DeleteMapper delete mapper from entity.
func (m *EntityManager) RemoveMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	if len(en.Mappers) == 0 {
		log.Errorf("append mapper into entity failed, %s", errEmptyEntityMapper)
		return nil, errors.Wrap(errEmptyEntityMapper, "append entity mapper failed")
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
	return en, nil
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
