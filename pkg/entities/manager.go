package entities

import (
	"context"
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
				entityInst, has := m.entities[msgCtx.Headers.GetTargetID()]
				if !has {
					// rebalance entity.
					if err := m.rebalanceEntity(context.Background(), msgCtx); nil != err {
						log.Errorf("dispose message failed, err: %s", err.Error())
					}
					continue
				}

				if entityInst.OnMessage(msgCtx.Message) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(entityInst.HandleLoop)
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

func (m *EntityManager) rebalanceEntity(ctx context.Context, msgCtx statem.MessageContext) error {
	// 1. 通过placement查询entity是否在当前节点.
	// 2.1. 如果在当前节点则在当前节点创建该实体.
	// 2.2 如果实体不属于当前节点，则将消息转发出去.

	var (
		err        error
		entityInst EntityOp
	)

	if false {
		// 从状态存储中获取实体信息.
	} else {
		entityInst, err = statem.NewState(context.Background(), m, &statem.Base{
			ID:    msgCtx.Headers.GetTargetID(),
			Type:  msgCtx.Headers.GetStateType(),
			Owner: msgCtx.Headers.GetOwner(),
		}, nil)
	}

	if nil != err {
		return errors.Wrap(err, "rrebalance entity failed")
	}

	entityInst.OnMessage(msgCtx.Message)
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
	panic("implement me.")
}

// GetProperties returns statem.Base.
func (m *EntityManager) GetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	panic("implement me.")
}

// SetProperties set properties into entity.
func (m *EntityManager) SetProperties(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	panic("implement me.")
}

// AppendMapper append a mapper into entity.
func (m *EntityManager) AppendMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	panic("implement me.")
}

// DeleteMapper delete mapper from entity.
func (m *EntityManager) DeleteMapper(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	panic("implement me.")
}
