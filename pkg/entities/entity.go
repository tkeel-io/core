package entities

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/tkeel-io/core/lock"
	"github.com/tkeel-io/core/pkg/mapper"
)

type EntityBase struct {
	ID      string                 `json:"id"`
	Tag     *string                `json:"tag"`
	Type    string                 `json:"type"`
	Source  string                 `json:"source"`
	UserID  string                 `json:"user_id"`
	Version int64                  `json:"version"`
	KValues map[string]interface{} `json:"properties"` //nolint
}

type entity struct {
	EntityBase

	mappers        map[string]mapper.Mapper          // key=mapperId
	tentacles      map[string][]mapper.Tentacler     // key=entityId#propertyKey
	cacheProps     map[string]map[string]interface{} // cache other property.
	indexTentacles map[string][]mapper.Tentacler     // key=targetId(mapperId/entityId)

	entityManager *EntityManager
	lock          *lock.ReEntryLock

	ctx context.Context
}

// newEntity create an entity object.
func newEntity(ctx context.Context, mgr *EntityManager, id string, source string, userID string, tag *string, version int64) (*entity, error) {
	if id == "" {
		id = uuid()
	}

	et := &entity{
		EntityBase: EntityBase{
			ID:      id,
			Tag:     tag,
			Source:  source,
			UserID:  userID,
			Version: version,
			KValues: make(map[string]interface{}),
		},

		ctx:           ctx,
		entityManager: mgr,
		lock:          lock.NewReEntryLock(0),
		mappers:       make(map[string]mapper.Mapper),
		cacheProps:    make(map[string]map[string]interface{}),
	}

	// set KValues into cacheProps.
	et.cacheProps[id] = et.KValues

	return et, nil
}

// GetID returns entity's id.
func (e *entity) GetID() string {
	return e.ID
}

// GetMapper returns a mapper.
func (e *entity) GetMapper(mid string) mapper.Mapper {
	reqID := uuid()

	log.Infof("entity.GetMapper called, entityId: %s, requestId: %s, mapperId: %s.", e.ID, reqID, mid)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

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

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()
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

	e.lock.Lock(&reqID)
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

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	return e.KValues[key]
}

// SetProperty set entity property.
func (e *entity) SetProperty(key string, value interface{}) error {
	reqID := uuid()

	log.Infof("entity.SetProperty called, entityId: %s, requestId: %s, key: %s.", e.ID, reqID, key)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	e.KValues[key] = value

	return nil
}

// GetAllProperties returns entity properties.
func (e *entity) GetAllProperties() *EntityBase {
	reqID := uuid()

	log.Infof("entity.GetAllProperties called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	return e.getEntityBase()
}

// SetProperties set entity properties.
func (e *entity) SetProperties(entityObj *EntityBase) (*EntityBase, error) {
	reqID := uuid()

	log.Infof("entity.SetProperties called, entityId: %s, requestId: %s.", e.ID, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	e.setTag(entityObj.Tag)

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

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	delete(e.KValues, key)

	return nil
}

func (e *entity) InvokeMsg(ctx EntityContext) {
	reqID := uuid()
	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	switch msg := ctx.Message.(type) {
	case *EntityMsg:
		e.invokeEntityMsg(msg)
	case *TentacleMsg:
		e.invokeTentacleMsg(msg)
	default:
		// invalid msg type.
		log.Errorf("undefine message type, msg: %s", msg)
	}
}

func (e *entity) invokeEntityMsg(msg *EntityMsg) {
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
			Message: &EntityMsg{
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
	ce := e.Copy()
	return &ce.EntityBase
}

func (e *entity) setTag(tag *string) {
	if nil != tag {
		tagVal := *tag
		e.Tag = &tagVal
	}
}

func (e entity) Copy() entity {
	return entity{
		EntityBase: EntityBase{
			ID:      e.ID,
			Tag:     e.Tag,
			Type:    e.Type,
			Source:  e.Source,
			UserID:  e.UserID,
			Version: e.Version,
			KValues: e.KValues,
		},
		mappers:        e.mappers,
		tentacles:      e.tentacles,
		cacheProps:     e.cacheProps,
		indexTentacles: e.indexTentacles,
		entityManager:  e.entityManager,
		lock:           e.lock,
		ctx:            e.ctx,
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
