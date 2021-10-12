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
		var et *entity
		if et = e.getEntity(entityId); nil == et {
			log.Warnf("entity not exists, entityId: %s.", entityId)
		}
		et.TentacleModify(reqID, e.Id)
	}

	return nil
}

// TentacleModify notify tentacle event.
func (e *entity) TentacleModify(requestId, entityId string) {

	tentacles := e.getEntity(entityId).GetTentacles(requestId, e.Id)

	e.lock.Lock(&requestId)
	defer e.lock.Unlock()

	if e.Id != entityId {
		e.indexTentacles[entityId] = tentacles
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

// GetTentacles returns tentacles.
func (e *entity) GetTentacles(requestId, entityId string) []mapper.Tentacler {

	e.lock.Lock(&requestId)
	defer e.lock.Unlock()

	tentacles := e.indexTentacles[entityId]
	result := make([]mapper.Tentacler, len(tentacles))

	for index, tentacle := range tentacles {
		result[index] = tentacle.Copy()
	}

	return result
}

// GetProperty returns entity property.
func (e *entity) GetProperty(key string) interface{} {

	e.lock.Lock(&requestId)
	defer e.lock.Unlock()

	return e.kvalues[e.Id][key]
}

// SetProperty set entity property.
func (e *entity) SetProperty(string, interface{}) error {
	panic("implement me.")
}

//GetAllProperties returns entity properties.
func (e *entity) GetAllProperties() map[string]interface{} {
	panic("implement me.")
}

// SetProperties set entity properties
func (e *entity) SetProperties(map[string]interface{}) error {
	panic("implement me.")
}

// DeleteProperty delete entity property.
func (e *entity) DeleteProperty(string) error {
	panic("implement me.")
}

// InvokeMsg
func (e *entity) InvokeMsg(entityId string, values map[string]interface{}) {
	panic("implement me.")
}

func (e *entity) getEntity(id string) *entity {
	return e.entityManager.GetEntity(id)
}
