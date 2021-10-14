package entities

import (
	"context"

	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/utils"
)

type entity struct {
	Id      string                            `json:"id"`
	Tag     string                            `json:"tag"`
	Source  string                            `json:"source"`
	UserId  string                            `json:"user_id"`
	Version int64                             `json:"version"`
	KValues map[string]map[string]interface{} `json:"kvalues"`

	mappers        map[string]mapper.Mapper      //key=mapperId
	tentacles      map[string][]mapper.Tentacler //key=entityId#propertyKey
	indexTentacles map[string][]mapper.Tentacler //key=targetId(mapperId/entityId)

	entityManager *manager
	lock          *utils.ReEntryLock

	ctx context.Context
}

// NewEntity create a entity object.
func NewEntity(ctx context.Context, mgr *manager, entityId string, source string, userId string, tag string, version int64) (*entity, error) {

	if entityId == "" {
		entityId = utils.GenerateUUID()
	}

	et := &entity{
		ctx:           ctx,
		Tag:           tag,
		Id:            entityId,
		Source:        source,
		UserId:        userId,
		Version:       version,
		entityManager: mgr,
		lock:          utils.NewReEntryLock(0),
		mappers:       make(map[string]mapper.Mapper),
		KValues:       make(map[string]map[string]interface{}),
	}

	//default this properties.
	et.KValues[entityId] = make(map[string]interface{})

	return et, mgr.Load(et)
}

// GetId returns entity's id.
func (e *entity) GetId() string {
	return e.Id
}

// GetMapper returns a mapper.
func (e *entity) GetMapper(mid string) mapper.Mapper {

	reqID := utils.GenerateUUID()
	log.Infof("entity.GetMapper called, entityId: %s, requestId: %s, mapperId: %s.",
		e.Id, reqID, mid)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	if m, _ := e.mappers[mid]; m != nil {
		return m.Copy()
	}

	return nil
}

// GetMappers
func (e *entity) GetMappers() []mapper.Mapper {

	var (
		result []mapper.Mapper
		reqID  = utils.GenerateUUID()
	)

	log.Infof("entity.GetMappers called, entityId: %s, requestId: %s.",
		e.Id, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	for _, m := range e.mappers {
		result = append(result, m.Copy())
	}

	return result
}

// SetMapper
func (e *entity) SetMapper(m mapper.Mapper) error {

	reqID := utils.GenerateUUID()
	log.Infof("entity.SetMapper called, entityId: %s, requestId: %s, mapperId: %s, mapper: %s.",
		e.Id, reqID, m.Id, m.String())

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	e.mappers[m.Id()] = m

	// generate indexTentacles again.
	for _, mp := range e.mappers {
		tentacles := m.Tentacles()
		for _, tentacle := range tentacles {
			e.indexTentacles[mp.TargetEntity()] =
				append(e.indexTentacles[mp.TargetEntity()], tentacle)
		}
	}

	sourceEntities := m.SourceEntities()
	for _, entityId := range sourceEntities {

		tentacle := mapper.MergeTentacles(e.indexTentacles[entityId]...)

		if nil != tentacle {
			// send tentacle msg.
			e.entityManager.SendMsg(EntityContext{
				Headers: Header{
					EntityCtxHeaderSourceId: e.Id,
					EntityCtxHeaderTargetId: entityId,
				},
				Message: &TentacleMsg{
					TargetId: e.Id,
					Items:    tentacle.Items(),
				},
			})
		}
	}

	return nil
}

// GetProperty returns entity property.
func (e *entity) GetProperty(key string) interface{} {

	reqID := utils.GenerateUUID()
	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s, key: %s.",
		e.Id, reqID, key)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	return e.KValues[e.Id][key]
}

// SetProperty set entity property.
func (e *entity) SetProperty(key string, value interface{}) error {
	reqID := utils.GenerateUUID()
	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s, key: %s.",
		e.Id, reqID, key)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	thisEntityProps := e.KValues[e.Id]
	thisEntityProps[key] = value

	return nil
}

//GetAllProperties returns entity properties.
func (e *entity) GetAllProperties() map[string]interface{} {
	reqID := utils.GenerateUUID()
	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s.",
		e.Id, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	props := e.KValues[e.Id]
	result := make(map[string]interface{})
	if err := utils.DeepCopy(&result, &props); nil != err {
		log.Errorf("duplicate properties failed.")
	}
	return result
}

// SetProperties set entity properties
func (e *entity) SetProperties(values map[string]interface{}) error {

	reqID := utils.GenerateUUID()
	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s.", e.Id, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	thisEntityProps := e.KValues[e.Id]

	for key, value := range values {
		thisEntityProps[key] = value
	}

	return nil
}

// DeleteProperty delete entity property.
func (e *entity) DeleteProperty(key string) error {

	reqID := utils.GenerateUUID()
	log.Infof("entity.GetProperty called, entityId: %s, requestId: %s.", e.Id, reqID)

	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	thisEntityProps := e.KValues[e.Id]
	delete(thisEntityProps, key)

	return nil
}

// InvokeMsg
func (e *entity) InvokeMsg(ctx EntityContext) {

	reqID := utils.GenerateUUID()
	e.lock.Lock(&reqID)
	defer e.lock.Unlock()

	switch msg := ctx.Message.(type) {
	case *EntityMsg:
		e.invokeEntityMsg(msg)
	case *TentacleMsg:
		e.invokeTentacleMsg(msg)
	default:
		//invalid msg type.
		log.Errorf("undefine message type, msg: %s", msg)
	}
}

//--------------考虑当entity自己的属性映射自己的时候

func (e *entity) invokeEntityMsg(msg *EntityMsg) {

	setEntityId := msg.SourceId
	if "" == setEntityId {
		setEntityId = e.Id
	}

	//1. update it's self properties.
	//2. generate message, then send msg.
	//3. active mapper.
	activeTentacles := make([]activePair, 0)
	entityProps := e.KValues[setEntityId]
	for key, value := range msg.Values {
		entityProps[key] = value
		activeTentacles = append(activeTentacles, activePair{key, mapper.GenTentacleKey(e.Id, key)})
	}

	//active tentacles.
	e.activeTentacle(activeTentacles)

}

func (e *entity) activeTentacle(actives []activePair) {

	var (
		activeMappers = make([]string, 0)
		messages      = make(map[string]map[string]interface{})
	)

	thisEntityProps := e.KValues[e.Id]
	for _, active := range actives {
		if tentacles, exists := e.tentacles[active.TentacleKey]; exists {
			for _, tentacle := range tentacles {
				targetId := tentacle.TargetId()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeMappers = append(activeMappers, targetId)
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					//make if not exists.
					if _, exists := messages[targetId]; exists {
						messages[targetId] = make(map[string]interface{})
					}

					// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
					messages[targetId][active.PropertyKey] = thisEntityProps[active.PropertyKey]
				} else {
					//undefine tentacle type.
					log.Warnf("undefine tentacle type, %v", tentacle)
				}
			}
		}
	}

	//send msgs.
	for entityId, msg := range messages {
		e.entityManager.SendMsg(EntityContext{
			Headers: Header{
				EntityCtxHeaderSourceId: e.Id,
				EntityCtxHeaderTargetId: entityId,
			},
			Message: &EntityMsg{
				SourceId: e.Id,
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

	if e.Id != msg.TargetId {

		tentacle := mapper.NewTentacle(mapper.TentacleTypeEntity, msg.TargetId, msg.Items)
		e.indexTentacles[msg.TargetId] = []mapper.Tentacler{tentacle}
	}

	//generate tentacles again.
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
