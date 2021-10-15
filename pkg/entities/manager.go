package entities

import (
	"context"
	"sync"

	ants "github.com/panjf2000/ants/v2"
)

type EntityManager struct {
	entities      map[string]*entity
	msgCh         chan EntityContext
	disposeCh     chan EntityContext
	coroutinePool *ants.Pool

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(ctx context.Context, coroutinePool *ants.Pool) *EntityManager {
	ctx, cancel := context.WithCancel(ctx)

	return &EntityManager{
		ctx:           ctx,
		cancel:        cancel,
		entities:      make(map[string]*entity),
		msgCh:         make(chan EntityContext),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}
}

func (m *EntityManager) Load(e *entity) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.entities[e.ID]; ok {
		return errEntityExisted
	}

	m.entities[e.ID] = e

	return nil
}

func (m *EntityManager) getEntity(id string) *entity {
	m.lock.Lock()
	defer m.lock.Unlock()

	if entityInst, ok := m.entities[id]; ok {
		return entityInst
	}
	return nil
}

func (m *EntityManager) GetProperty(ctx context.Context, entityID, propertyKey string) (resp interface{}, err error) {
	m.lock.Lock()
	entityInst, has := m.entities[entityID]
	m.lock.Unlock()

	if !has {
		err = errEntityNotFound
		log.Errorf("EntityManager.GetProperty failed, err: %s", err.Error())
	}

	return entityInst.GetProperty(propertyKey), err
}

func (m *EntityManager) GetAllProperties(ctx context.Context, entityID string) (resp interface{}, err error) {
	m.lock.Lock()
	entityInst, has := m.entities[entityID]
	m.lock.Unlock()

	if !has {
		err = errEntityNotFound
		log.Errorf("EntityManager.GetAllProperties failed, err: %s", err.Error())
	}

	return entityInst.GetAllProperties(), err
}

func (m *EntityManager) SetProperties(ctx context.Context, entityObj *EntityBase) error {
	m.lock.Lock()
	entityInst, has := m.entities[entityObj.ID]
	// check id, source, userId, ...
	m.lock.Unlock()

	if !has {
		err := errEntityNotFound
		log.Errorf("EntityManager.GetAllProperties failed, err: %s", err.Error())
	}

	// 对于Header同步落盘，对于Kvalues 延迟落盘，可以做缓冲，罗盘策略：定时+定量。

	entityInst.SetProperties(entityObj.KValues)

	return nil
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

	for key := range entityInst.KValues {
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
				m.coroutinePool.Submit(func() {
					if entity, has := m.entities[entityCtx.TargetID()]; has {
						entity.InvokeMsg(entityCtx)
					} else {
						log.Warnf("dispose msg failed, entity(%s) not found.", entityCtx.TargetID())
					}
				})
			}
		}
	}()

	return nil
}

func init() {}
