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

	lock sync.RWMutex
	ctx  context.Context
}

func NewEntityManager(ctx context.Context, coroutinePool *ants.Pool) *EntityManager {

	return &EntityManager{
		ctx:           ctx,
		entities:      make(map[string]*entity),
		msgCh:         make(chan EntityContext), //在channel内部传递引用或指针可能造成gc回收困难和延迟。
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}
}

func (m *EntityManager) Load(e *entity) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.entities[e.Id]; ok {
		return errEntityExisted
	}

	m.entities[e.Id] = e

	return nil
}

func (m *EntityManager) GetEntity(id string) *entity {

	m.lock.Lock()
	defer m.lock.Unlock()

	entity, _ := m.entities[id]
	return entity
}

func (m *EntityManager) GetProperty(ctx context.Context, entityId, propertyKey string) (resp interface{}, err error) {

	m.lock.Lock()
	entity, has := m.entities[entityId]
	m.lock.Unlock()

	if !has {
		err = errEntityNotFound
		log.Errorf("EntityManager.GetProperty failed, err: %s", err.Error())
	}

	return entity.GetProperty(propertyKey), err
}

func (m *EntityManager) SendMsg(ctx EntityContext) {
	//解耦actor之间的直接调用

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
				//dispatch message. 将消息分发到不同的节点。
				m.disposeCh <- entityCtx

			case entityCtx := <-m.disposeCh:
				//invoke msg.
				m.coroutinePool.Submit(func() {
					if entity, has := m.entities[entityCtx.TargetId()]; has {
						entity.InvokeMsg(entityCtx)
					} else {
						log.Warnf("dispose msg failed, entity(%s) not found.", entityCtx.TargetId())
					}
				})
			}
		}
	}()

	return nil
}

func init() {}
