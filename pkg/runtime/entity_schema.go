package runtime

import (
	"github.com/tkeel-io/tdtl"
	"sync"
)

type NodeCache struct {
	cache      map[string]map[string]tdtl.Node
	schemeLock sync.RWMutex
}

func NewNodeCache() *NodeCache {
	return &NodeCache{
		cache: make(map[string]map[string]tdtl.Node),
	}
}

func (nc *NodeCache) Get(entity, path string) (tdtl.Node, bool) {
	nc.schemeLock.RLock()
	defer nc.schemeLock.RUnlock()
	if _, ok := nc.cache[entity]; ok {
		if node, ok := nc.cache[entity][path]; ok {
			return node, true
		}
	}
	return tdtl.UNDEFINED_RESULT, false
}

func (nc *NodeCache) Set(entity, path string, node tdtl.Node) {
	nc.schemeLock.Lock()
	defer nc.schemeLock.Unlock()
	if _, ok := nc.cache[entity]; !ok {
		nc.cache[entity] = make(map[string]tdtl.Node)
	}
	nc.cache[entity][path] = node
}

func (nc *NodeCache) Delete(entity string) {
	nc.schemeLock.Lock()
	defer nc.schemeLock.Unlock()
	delete(nc.cache, entity)
}
