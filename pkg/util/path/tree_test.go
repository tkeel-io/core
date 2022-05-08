package path

import (
	"sort"
	"fmt"
	"strings"
	"testing"
)

func TestTree(t *testing.T) {
	tr := New()
	addNode := func(path string) {
		path += ".*"
		tr.Add(path, &NodeMock{path: path})
	}
	delNode := func(path string) {
		path += ".*"
		tr.Remove(path, &NodeMock{path: path})
	}
	// 需要订阅或执行的语句
	addNode("a.x[0]")
	addNode("a")
	addNode("a.b.c")
	addNode("a.b.c.d")
	addNode("a.b.c.d")
	addNode("a.b.c.d")
	delNode("a.b.c.d")
	addNode("a.b.c.d.e")
	addNode("a.b.c.d.e.f")
	addNode("a.b.c.d.e.i")
	addNode("a.b.c.d.e.g")
	addNode("a.b.c.d.e.i.j")
	addNode("a.b.c.d.e.i.j.k")

	fmt.Println(tr.String())

	tests := []struct {
		name string
		path string
		want string
	}{
		{"1", "a.b.c.r", "a.*|a.b.c.*"},
		{"2", "a", "a.*|a.b.c.*|a.b.c.d.e.*|a.b.c.d.e.f.*|a.b.c.d.e.g.*|a.b.c.d.e.i.*|a.b.c.d.e.i.j.*|a.b.c.d.e.i.j.k.*|a.x[0].*"},
		{"3", "b.b.c.r", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			ret := []string{}
			nodes := tr.MatchPrefix(tt.path)
			for _, node := range nodes {
				ret = append(ret, node.String())
			}
			sort.Strings(ret)
			size := len(ret)
			got := strings.Join(ret, "|")
			if got != tt.want {
				t1.Errorf("Add() = [%d]%v, want %v", size, got, tt.want)
			}
		})
	}
}

func TestRefTree(t *testing.T) {
	tr := NewRefTree()
	addNode := func(path string) {
		path += ".*"
		tr.Add(path, &NodeMock{path: path})
	}
	delNode := func(path string) {
		path += ".*"
		tr.Remove(path, &NodeMock{path: path})
	}
	// 需要订阅或执行的语句
	addNode("a.x[0]")
	addNode("a")
	addNode("a.b.c")
	addNode("a.b.c.d")
	addNode("a.b.c.d")
	addNode("a.b.c.d")
	delNode("a.b.c.d")
	addNode("a.b.c.d.e")
	addNode("a.b.c.d.e.f")
	addNode("a.b.c.d.e.i")
	addNode("a.b.c.d.e.g")
	addNode("a.b.c.d.e.i.j")
	addNode("a.b.c.d.e.i.j.k")

	tests := []struct {
		name string
		path string
		want string
	}{
		{"1", "a.b.c.r", "a.*|a.b.c.*"},
		{"2", "a", "a.*|a.b.c.*|a.b.c.d.*|a.b.c.d.e.*|a.b.c.d.e.f.*|a.b.c.d.e.g.*|a.b.c.d.e.i.*|a.b.c.d.e.i.j.*|a.b.c.d.e.i.j.k.*|a.x[0].*"},
		{"1", "b.b.c.r", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			ret := []string{}
			nodes := tr.MatchPrefix(tt.path)
			for _, node := range nodes {
				ret = append(ret, node.String())
			}
			sort.Strings(ret)
			size := len(ret)
			got := strings.Join(ret, "|")
			if got != tt.want {
				t1.Errorf("Add() = [%d]%v, want %v", size, got, tt.want)
			}
		})
	}
}
