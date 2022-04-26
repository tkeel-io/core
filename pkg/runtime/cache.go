package runtime

import (
	"context"

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/kit/log"
)

type EntityCache interface {
	Load(ctx context.Context, id string) (Entity, error)
	Snapshot() error
}

type eCache struct {
	entities   map[string]Entity
	repository repository.IRepository
}

func NewCache(repo repository.IRepository) EntityCache {
	return &eCache{repository: repo,
		entities: make(map[string]Entity)}
}

func (ec *eCache) Load(ctx context.Context, id string) (Entity, error) {
	if state, ok := ec.entities[id]; ok {
		return state, nil
	}

	// load from state storage.
	jsonData, err := ec.repository.GetEntity(context.TODO(), id)
	if nil != err {
		log.L().Warn("load entity from state storage",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "load entity")
	}

	// create entity instance.
	en, err := NewEntity(id, jsonData)
	if nil != err {
		log.L().Warn("create entity instance",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "create entity instance")
	}

	// cache entity.
	ec.entities[id] = en
	return en, errors.Wrap(err, "load cache entity")
}

func (ec *eCache) Snapshot() error {
	panic("implement me")
}
