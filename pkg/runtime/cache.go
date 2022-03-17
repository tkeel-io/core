package runtime

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

type EntityCache interface {
	Load(id string) (Entity, error)
	Snapshot() error
}

type eCache struct {
	entities map[string]Entity
}

func NewCache() EntityCache {
	return &eCache{}
}

func (ec *eCache) Load(id string) (Entity, error) {
	if state, ok := ec.entities[id]; ok {
		return state, nil
	}

	// load from state store.
	cc := tdtl.New([]byte(`{"properties":{}}`))
	cc.Set("id", tdtl.New(id))
	en, err := NewEntity(id, cc.Raw())
	if nil == err {
		// cache entity.
		ec.entities[id] = en
	}
	return en, errors.Wrap(err, "load cache entity")
}

func (ec *eCache) Snapshot() error {
	panic("implement me")
}
