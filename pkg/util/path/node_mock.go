package path

type NodeMock struct {
	path string
}

func (m *NodeMock) ID() string {
	return m.path
}

func (m *NodeMock) String() string {
	return m.path
}
