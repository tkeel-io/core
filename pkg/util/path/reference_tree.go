package path

import (
	"strings"
	"sync"
)

type RefNode struct {
	node    Node
	counter int
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
	defer rt.mutex.Unlock()
	rt.mutex.Lock()
	if _, exists := rt.nodes[value.ID()]; !exists {
		rt.nodes[value.ID()] = &RefNode{node: value}
	}
	rt.nodes[value.ID()].Inc()

	// replace node.
	return rt.tree.add(value, 0, strings.Split(fmtPath(path), rt.tree.Separator), rt.tree.root)
}

func (rt *RefTree) Remove(path string, value Node) bool {
	// increase counter if node exists.
	defer rt.mutex.Unlock()
	rt.mutex.Lock()
	if _, exists := rt.nodes[value.ID()]; !exists {
		return false
	}

	// decrease node counter.
	flag := rt.nodes[value.ID()].Dec()
	if flag {
		delete(rt.nodes, value.ID())
	}

	// remove node.
	if flag {
		return rt.tree.remove(value, 0, strings.Split(fmtPath(path), rt.tree.Separator), rt.tree.root)
	}

	return true
}

func (rt *RefTree) MatchPrefix(path string) []Node {
	return rt.tree.MatchPrefix(path)
}

func (rt *RefTree) String() string {
	return rt.tree.String()
}
