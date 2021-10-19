package entities

import (
	"context"
	"fmt"
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
		msgCh:         make(chan EntityContext, 10),
		disposeCh:     make(chan EntityContext, 10),
		coroutinePool: coroutinePool,
		lock:          sync.RWMutex{},
	}
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

func (m *EntityManager) checkEntity(ctx context.Context, entityObj *EntityBase) (entityInst *entity, err error) {
	var (
		has         bool
		emptyString = ""
	)

	if entityInst, has = m.entities[entityObj.ID]; has {
		return
	}

	// require Type, UserId, Source.
	if emptyString == entityObj.Type {
		err = entityFieldRequired("Type")
	} else if emptyString == entityObj.UserID {
		err = entityFieldRequired("UserId")
	} else if emptyString == entityObj.Source {
		err = entityFieldRequired("Source")
	} else {
		entityInst, err = newEntity(context.Background(), m, entityObj.ID, entityObj.Source, entityObj.UserID, entityObj.Tag, 0)
		if nil == err {
			m.entities[entityInst.ID] = entityInst
		}
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

	return entityInst.SetProperties(entityObj)
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

func (m *EntityManager) HandleMsg() {
	// dispose message from pubsub.

}

func init() {}
