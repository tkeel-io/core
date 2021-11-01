package entities

import (
	"context"
	"fmt"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
)

type EntityManager struct {
	entities      map[string]EntityOp
	msgCh         chan EntityContext
	disposeCh     chan EntityContext
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
		msgCh:         make(chan EntityContext, 10),
		disposeCh:     make(chan EntityContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}, nil
}

func (m *EntityManager) DeleteEntity(ctx context.Context, entityObj *EntityBase) (*EntityBase, error) {
	m.lock.Lock()
	entityInst, has := m.entities[entityObj.ID]
	delete(m.entities, entityObj.ID)
	m.lock.Unlock()

	if !has {
		log.Errorf("EntityManager.GetAllProperties failed, entity not found.")
		return nil, errEntityNotFound
	}

	entityObj = entityInst.GetAllProperties()

	return entityObj, nil
}

func (m *EntityManager) GetProperty(ctx context.Context, entityID, propertyKey string) (resp interface{}, err error) {
	m.lock.Lock()
	entityInst, has := m.entities[entityID]
	m.lock.Unlock()

	if !has {
		log.Errorf("EntityManager.GetAllProperties failed, entity not found.")
		return nil, errEntityNotFound
	}

	return entityInst.GetProperty(propertyKey), err
}

func (m *EntityManager) GetAllProperties(ctx context.Context, entityObj *EntityBase) (*EntityBase, error) {
	m.lock.Lock()
	entityInst, has := m.entities[entityObj.ID]
	m.lock.Unlock()

	if !has {
		log.Errorf("EntityManager.GetAllProperties failed, entity not found.")
		return nil, errEntityNotFound
	}

	return entityInst.GetAllProperties(), nil
}

func (m *EntityManager) checkEntity(ctx context.Context, entityObj *EntityBase) (entityInst EntityOp, err error) {
	var has bool

	if entityInst, has = m.entities[entityObj.ID]; has {
		return
	}

	// create entity.
	switch entityObj.Type {
	case EntityTypeState:
		entityInst, err = newEntity(context.Background(), m, entityObj)
	case EntityTypeDevice:
		entityInst, err = newEntity(context.Background(), m, entityObj)
	case EntityTypeSpace:
	case EntityTypeSubscription:
		entityInst, err = newSubscription(context.Background(), m, entityObj)
	default:
		entityInst, err = newEntity(context.Background(), m, entityObj)
	}

	if nil == err {
		m.entities[entityInst.GetID()] = entityInst
	}

	return
}

func (m *EntityManager) SetProperties(ctx context.Context, entityObj *EntityBase) (*EntityBase, error) {
	m.lock.Lock()
	// check id, type, source, userId.
	entityInst, err := m.checkEntity(ctx, entityObj)
	m.lock.Unlock()

	if nil != err {
		return nil, fmt.Errorf("entityManager.SetProperties failed, %w", err)
	}

	if len(entityObj.Mappers) > 0 {
		if err = entityInst.SetMapper(entityObj.Mappers[0]); nil != err {
			return nil, errors.Wrap(err, "entityManager.SetProperties failed")
		}
	}

	entityObj, err = entityInst.SetProperties(entityObj)

	return entityObj, errors.Wrap(err, "set properties failed")
}

func (m *EntityManager) DeleteProperty(ctx context.Context, entityObj *EntityBase) error {
	m.lock.Lock()
	entityInst, has := m.entities[entityObj.ID]
	// check id, source, userId, ...
	m.lock.Unlock()

	if !has {
		err := errEntityNotFound
		log.Errorf("EntityManager.GetAllProperties failed, err: %s", err.Error())
	}

	for key := range entityObj.KValues {
		entityInst.DeleteProperty(key)
	}

	return nil
}

func (m *EntityManager) SendMsg(ctx EntityContext) {
	// 解耦actor之间的直接调用
	m.msgCh <- ctx
}

func (m *EntityManager) Start() error {
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				log.Info("entity EntityManager exited.")
				return
			case entityCtx := <-m.msgCh:
				// dispatch message. 将消息分发到不同的节点。
				m.disposeCh <- entityCtx

			case entityCtx := <-m.disposeCh:
				// invoke msg.
				// 实际上reactor模式有一个致命的问题就是消息乱序, 引入mailbox可以有效规避乱序问题.
				var err error
				entityInst, has := m.entities[entityCtx.Headers.GetTargetID()]
				if !has {
					m.lock.Lock()
					entityObj := &EntityBase{
						ID:       entityCtx.Headers.GetTargetID(),
						Type:     entityCtx.Headers.GetEntityType(),
						Owner:    entityCtx.Headers.GetOwner(),
						PluginID: entityCtx.Headers.GetPluginID(),
					}
					entityInst, err = m.checkEntity(context.TODO(), entityObj)
					m.lock.Unlock()
				}

				if nil != err {
					log.Warnf("dispose msg failed, entity(%s) not found.", entityCtx.Headers.GetTargetID())
					continue
				}

				if entityInst.OnMessage(entityCtx) {
					// attatch goroutine to entity.
					m.coroutinePool.Submit(entityInst.InvokeMsg)
				}
			}
		}
	}()

	return nil
}

func (m *EntityManager) HandleMsg(ctx context.Context, msg EntityContext) {
	// dispose message from pubsub.
	m.msgCh <- msg
}

func (m *EntityManager) EscapedEntities(expression string) []string {
	entities := []string{}
	switch expression {
	case "*":
		m.lock.RLock()
		for entityID := range m.entities {
			entities = append(entities, entityID)
		}
		m.lock.RUnlock()
	default:
		entities = []string{expression}
	}

	return entities
}

func init() {}
