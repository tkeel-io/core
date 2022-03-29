package placement

import (
	"sort"
	"sync"

	"github.com/tkeel-io/core/pkg/util"
)

var globalPlacement *placement

type placement struct {
	lock      sync.RWMutex
	queues    map[string]Info
	hashTable sort.StringSlice
}

func New() Placement {
	return &placement{}
}

func (p *placement) Append(info Info) {
	p.lock.Lock()
	p.queues[info.ID] = info
	p.hashTable = append(p.hashTable, info.ID)
	sort.Sort(p.hashTable)
	p.hashTable = util.Unique(p.hashTable)
	p.lock.Unlock()
}

func (p *placement) Remove(info Info) {
	p.lock.Lock()
	delete(p.queues, info.ID)
	index := p.hashTable.Search(info.ID)
	if p.hashTable.Len() > index && info.ID == p.hashTable[index] {
		p.hashTable = append(p.hashTable[:index], p.hashTable[index+1:]...)
	}
	p.lock.Unlock()
}

func (p *placement) Select(key string) Info {
	hashKey := util.Hash32(key)
	p.lock.RLock()
	selectIndex := hashKey % uint32(p.hashTable.Len())
	info := p.queues[p.hashTable[selectIndex]]
	p.lock.RUnlock()
	return info
}

func Initialize() {
	globalPlacement = &placement{
		lock:      sync.RWMutex{},
		queues:    make(map[string]Info),
		hashTable: sort.StringSlice{},
	}
}
