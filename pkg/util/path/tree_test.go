package path

import "testing"

func TestTree(t *testing.T) {
	tr := New()
	tr.Add("a.x[0]", &NodeMock{path: "1"})
	tr.Add("a.*", &NodeMock{path: "11"})
	tr.Add("a.b.c.d.e.f", &NodeMock{path: "2"})
	tr.Add("a.b.c.d.e.i", &NodeMock{path: "3"})
	tr.Add("a.b.c.d.e.i.j", &NodeMock{path: "13"})
	tr.Add("a.b.c.d.e.i.j.k", &NodeMock{path: "113"})
	tr.Add("a.b.c.d.e.g", &NodeMock{path: "4"})
	tr.Add("a.b.c.d.e.g", &NodeMock{path: "5"})
	tr.Add("a.b.c.d.e.g", &NodeMock{path: "6"})
	tr.Add("a.b.c.d.e.g", &NodeMock{path: "7"})

	tr.Print()

	nodes := tr.MatchPrefix("a.b.c.d.e")

	t.Log(nodes)
}
