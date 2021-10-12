package entities

import "sync"

type manager struct {
	entities map[string]*entity
	lock     sync.RWMutex
}

func NewManager() *manager {
	return &manager{
		entities: make(map[string]*entity),
		lock:     sync.RWMutex{},
	}
}

func (m *manager) Load(e *entity) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.entities[e.Id]; ok {
		return errEntityExisted
	}

	m.entities[e.Id] = e

	return nil
}

func (m *manager) GetEntity(id string) *entity {

	m.lock.Lock()
	defer m.lock.Unlock()

	entity, _ := m.entities[id]
	return entity
}

func init() {}
