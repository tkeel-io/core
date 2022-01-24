package placement

import (
	"sort"
	"sync"

	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
)

var globalPlacement *placement

type placement struct {
	lock      sync.RWMutex
	queues    map[string]dao.Queue
	hashTable sort.StringSlice
}

func New() Placement {
	return &placement{}
}

func (p *placement) AppendQueue(queue dao.Queue) {
	p.lock.Lock()
	p.queues[queue.ID] = queue
	p.hashTable = append(p.hashTable, queue.ID)
	sort.Sort(p.hashTable)
	p.lock.Unlock()
}

func (p *placement) RemoveQueue(queue dao.Queue) {
	p.lock.Lock()
	delete(p.queues, queue.ID)
	index := p.hashTable.Search(queue.ID)
	if p.hashTable.Len() > index && queue.ID == p.hashTable[index] {
		p.hashTable = append(p.hashTable[:index], p.hashTable[index+1:]...)
	}
	p.lock.Unlock()
}

func (p *placement) Select(key string) dao.Queue {
	hashKey := util.Hash32(key)
	p.lock.RLock()
	selectIndex := hashKey % uint32(p.hashTable.Len())
	selectQueue := p.queues[p.hashTable[selectIndex]]
	p.lock.RUnlock()
	return selectQueue
}

func Initialize() {
	globalPlacement = &placement{lock: sync.RWMutex{}}
}
