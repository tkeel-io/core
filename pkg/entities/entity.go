package entities

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tkeel-io/core/pkg/mapper"
)

const (
	EntityTypeState        = "STATE"
	EntityTypeDevice       = "DEVICE"
	EntityTypeSpace        = "SPACE"
	EntityTypeSubscription = "SUBSCRIPTION"

	EntityDisposingIdle  int32 = 0
	EntityDisposingSync  int32 = 1
	EntityDisposingAsync int32 = 2

	EntityDetached int32 = 0
	EntityAttached int32 = 1

	EntityStatusActive  = "active"
	EntityStatusDeleted = "deleted"
)

type EntityBase struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Owner    string                 `json:"owner"`
	Status   string                 `json:"status"`
	Version  int64                  `json:"version"`
	PluginID string                 `json:"plugin_id"`
	LastTime int64                  `json:"last_time"`
	KValues  map[string]interface{} `json:"properties"` //nolint
}

type entity struct {
	EntityBase

	mailBox   *mailbox
	disposing int32

	// non-state.
	mappers        map[string]mapper.Mapper          // key=mapperId
	tentacles      map[string][]mapper.Tentacler     // key=entityId#propertyKey
	cacheProps     map[string]map[string]interface{} // cache other property.
	indexTentacles map[string][]mapper.Tentacler     // key=targetId(mapperId/entityId)

	attached      int32
	entityManager *EntityManager
	lock          *sync.RWMutex

	ctx context.Context
}

// newEntity create an entity object.
func newEntity(ctx context.Context, mgr *EntityManager, in *EntityBase) (*entity, error) {
	if in.ID == "" {
		in.ID = uuid()
	}

	et := &entity{
		EntityBase: EntityBase{
			ID:       in.ID,
			Type:     in.Type,
			Owner:    in.Owner,
			PluginID: in.PluginID,
			Status:   EntityStatusActive,
			KValues:  make(map[string]interface{}),
		},

		ctx:           ctx,
		entityManager: mgr,
		mailBox:       newMailbox(10),
		lock:          &sync.RWMutex{},
		disposing:     EntityDisposingIdle,
		mappers:       make(map[string]mapper.Mapper),
		cacheProps:    make(map[string]map[string]interface{}),
	}

	// set KValues into cacheProps.
	et.cacheProps[in.ID] = et.KValues

	return et, nil
}

// GetMapper returns a mapper.
func (e *entity) GetMapper(mid string) mapper.Mapper {
	reqID := uuid()

	log.Infof("entity.GetMapper called, entityId: %s, requestId: %s, mapperId: %s.", e.ID, reqID, mid)

	e.lock.RLock()
	defer e.lock.RUnlock()

	if m, has := e.mappers[mid]; has {
		return m.Copy()
	}

	return nil
}

func (e *entity) GetMappers() []mapper.Mapper {
	var (
		result = make([]mapper.Mapper, len(e.mappers))
		reqID  = uuid()
	)

	log.Infof("entity.GetMappers called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.RLock()
	defer e.lock.RUnlock()
	i := 0
	for _, m := range e.mappers {
		result[i] = m.Copy()
		i++
	}

	return result
}

func (e *entity) SetMapper(m mapper.Mapper) error {
	reqID := uuid()

	log.Infof("entity.SetMapper called, entityID: %s, requestId: %s, mapperId: %s, mapper: %s.",
		e.ID, reqID, m.ID, m.String())

	e.lock.Lock()
	defer e.lock.Unlock()

	e.mappers[m.ID()] = m

	// generate indexTentacles again.
	for _, mp := range e.mappers {
		e.indexTentacles[mp.TargetEntity()] =
			append(e.indexTentacles[mp.TargetEntity()], m.Tentacles()...)
	}

	sourceEntities := m.SourceEntities()
	for _, entityID := range sourceEntities {
		tentacle := mapper.MergeTentacles(e.indexTentacles[entityID]...)

		if nil != tentacle {
			// send tentacle msg.
			e.entityManager.SendMsg(EntityContext{
				Headers: Header{
					EntityCtxHeaderSourceID: e.ID,
					EntityCtxHeaderTargetID: entityID,
				},
				Message: &TentacleMsg{
					TargetID: e.ID,
					Items:    tentacle.Items(),
				},
			})
		}
	}

	return nil
}

// GetProperty returns entity property.
func (e *entity) GetProperty(key string) interface{} {
	reqID := uuid()

	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s, key: %s.", e.ID, reqID, key)

	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.KValues[key]
}

// GetAllProperties returns entity properties.
func (e *entity) GetAllProperties() *EntityBase {
	reqID := uuid()

	log.Infof("entity.GetAllProperties called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.getEntityBase()
}

// SetProperties set entity properties.
func (e *entity) SetProperties(entityObj *EntityBase) (*EntityBase, error) {
	reqID := uuid()

	log.Infof("entity.SetProperties called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.Lock()
	defer e.lock.Unlock()

	e.Version++
	e.LastTime = time.Now().UnixNano() / 1e6

	// so dengerous, think Kvalues Store.
	for key, value := range entityObj.KValues {
		e.KValues[key] = value
	}

	return e.getEntityBase(), nil
}

// DeleteProperty delete entity property.
func (e *entity) DeleteProperty(key string) error {
	reqID := uuid()

	log.Infof("entity.DeleteProperty called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.KValues, key)

	return nil
}

func (e *entity) OnMessage(ctx EntityContext) bool {
	reqID := uuid()

	log.Infof("entity.OnMessage called, entityId: %s, requestId: %s.", e.ID, reqID)

	// 1. put msg inti mailbox.promise_handler
	// 2. start consume mailbox.
	attaching := false

	for {
		// 如果只有一条投递线程，那么会导致Dispatcher上的所有Entity都依赖于Message Queue中的消息的均匀性.
		if nil == e.mailBox.Put(ctx.Message) {
			break
		}
		runtime.Gosched()
	}

	if atomic.CompareAndSwapInt32(&e.attached, EntityDetached, EntityAttached) {
		attaching = true
		log.Infof("attatched entity, id: %s.", e.ID)
	}

	return attaching
}

func (e *entity) InvokeMsg() {
	for {
		var msgCtx Message
		if msgCtx = e.mailBox.Get(); nil == msgCtx {
			// detach this entity.
			if atomic.CompareAndSwapInt32(&e.attached, EntityAttached, EntityDetached) {
				log.Infof("detached entity, id: %s.", e.ID)
			}
			break
		}

		// lock messages.
		e.lock.Lock()

		switch msg := msgCtx.(type) {
		case *EntityMessage:
			e.invokeEntityMsg(msg)
		case *TentacleMsg:
			e.invokeTentacleMsg(msg)
		default:
			// invalid msg type.
			log.Errorf("undefine message type, msg: %s", msg)
		}

		e.lock.Unlock()
	}
}

func (e *entity) invokeEntityMsg(msg *EntityMessage) {
	setEntityID := msg.SourceID
	if setEntityID == "" {
		setEntityID = e.ID
	}

	// 1. update itself properties.
	// 2. generate message, then send msg.
	// 3. active mapper.
	activeTentacles := make([]activePair, 0)
	entityProps := e.cacheProps[setEntityID]
	for key, value := range msg.Values {
		entityProps[key] = value
		activeTentacles = append(activeTentacles, activePair{key, mapper.GenTentacleKey(e.ID, key)})
	}

	e.LastTime = time.Now().UnixNano() / 1e6
	// active tentacles.
	e.activeTentacle(activeTentacles)
}

func (e *entity) activeTentacle(actives []activePair) {
	var (
		activeMappers = make([]string, 0)
		messages      = make(map[string]map[string]interface{})
	)

	thisEntityProps := e.cacheProps[e.ID]
	for _, active := range actives {
		if tentacles, exists := e.tentacles[active.TentacleKey]; exists {
			for _, tentacle := range tentacles {
				targetID := tentacle.TargetID()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeMappers = append(activeMappers, targetID)
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					// make if not exists.
					if _, exists := messages[targetID]; exists {
						messages[targetID] = make(map[string]interface{})
					}

					// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
					messages[targetID][active.PropertyKey] = thisEntityProps[active.PropertyKey]
				} else {
					// undefined tentacle type.
					log.Warnf("undefined tentacle type, %v", tentacle)
				}
			}
		}
	}

	for entityID, msg := range messages {
		e.entityManager.SendMsg(EntityContext{
			Headers: Header{
				EntityCtxHeaderSourceID: e.ID,
				EntityCtxHeaderTargetID: entityID,
			},
			Message: &EntityMessage{
				SourceID: e.ID,
				Values:   msg,
			},
		})
	}

	// active mapper.
	e.activeMapper(activeMappers)
}

func (e *entity) activeMapper(actives []string) {

}

func (e *entity) invokeTentacleMsg(msg *TentacleMsg) {
	if e.ID != msg.TargetID {
		tentacle := mapper.NewTentacle(mapper.TentacleTypeEntity, msg.TargetID, msg.Items)
		e.indexTentacles[msg.TargetID] = []mapper.Tentacler{tentacle}
	}

	// generate tentacles again.
	e.tentacles = make(map[string][]mapper.Tentacler)
	for _, tentacles := range e.indexTentacles {
		for _, tentacle := range tentacles {
			for _, item := range tentacle.Items() {
				e.tentacles[item] = append(e.tentacles[item], tentacle)
			}
		}
	}
}

type activePair struct {
	PropertyKey string
	TentacleKey string
}

func (e *entity) getEntityBase() *EntityBase {
	return &EntityBase{
		ID:       e.ID,
		Type:     e.Type,
		Status:   e.Status,
		Owner:    e.Owner,
		Version:  e.Version,
		KValues:  e.KValues,
		PluginID: e.PluginID,
		LastTime: e.LastTime,
	}
}

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
