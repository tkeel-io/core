package service

import (
	"context"
	"strings"

	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/store"
)

const keyPrefix = "getEntityData_"

type EntityHistory interface {
	AddEnity(user, entityID string)
	GetLatestEntities(user string) []string
}

type entityHistory struct {
	count int
	store store.Store
}

func (e *entityHistory) AddEnity(user, entityID string) {
	item, err := e.store.Get(context.Background(), keyPrefix+user)
	entities := make([]string, 0)

	entities = append(entities, entityID)
	if err == nil {
		cacheMap := make(map[string]struct{})
		cacheMap[entityID] = struct{}{}
		oldCache := e.bytes2stringList(item.Value)

		count := 1
		for _, v := range oldCache {
			if _, ok := cacheMap[v]; ok {
				continue
			}
			entities = append(entities, v)
			cacheMap[v] = struct{}{}
			count++
			if count >= e.count {
				break
			}
		}
	}
	e.store.Set(context.Background(), keyPrefix+user, e.stringList2bytes(entities))
}

func (e *entityHistory) stringList2bytes(req []string) []byte {
	return []byte(strings.Join(req, ","))
}

func (e *entityHistory) bytes2stringList(req []byte) []string {
	return strings.Split(string(req), ",")
}

func (e *entityHistory) GetLatestEntities(user string) []string {
	item, err := e.store.Get(context.Background(), keyPrefix+user)
	if err != nil {
		return nil
	}
	return e.bytes2stringList(item.Value)
}

func NewEntityHistory(metadata resource.Metadata, maxcount int) EntityHistory {
	e := &entityHistory{
		count: maxcount,
	}
	e.store = store.NewStore(metadata)
	return e
}
