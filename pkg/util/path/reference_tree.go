package path

import (
	"sync"
)

type RefNode struct {
	node    Node
	counter int64
}

func (n *RefNode) Inc() bool {
	n.counter++
	return true
}

func (n *RefNode) Dec() bool {
	n.counter--
	return n.counter == 0
}

func NewRefTree() *RefTree {
	return &RefTree{
		tree:  New(),
		nodes: make(map[string]*RefNode),
	}
}

type RefTree struct {
	tree  *Tree
	mutex sync.RWMutex
	nodes map[string]*RefNode
}

func (rt *RefTree) Add(path string, value Node) bool {
	rt.mutex.Lock()
	if _, exists := rt.nodes[value.ID()]; !exists {
		rt.nodes[value.ID()] = &RefNode{node: value}
	}
	rt.nodes[value.ID()].Inc()
	rt.mutex.Unlock()

	// replace node.
	return rt.tree.Add(path, value)
}

func (rt *RefTree) Remove(path string, value Node) bool {
	// increase counter if node exists.
	rt.mutex.Lock()
	if _, exists := rt.nodes[value.ID()]; !exists {
		rt.mutex.Unlock()
		return false
	}

	// decrease node counter.
	flag := rt.nodes[value.ID()].Dec()
	if flag {
		delete(rt.nodes, value.ID())
	}

	rt.mutex.Unlock()

	// remove node.
	if flag {
		return rt.tree.Remove(path, value)
	}

	return true
}

func (rt *RefTree) MatchPrefix(path string) []Node {
	return rt.tree.MatchPrefix(path)
}
