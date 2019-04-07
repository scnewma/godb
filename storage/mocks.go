package storage

type MockStorage struct {
	GetFn func(string) (Node, error)
	SetFn func(string, Node)
	DelFn func(string) int
}

func (m *MockStorage) Get(key string) (Node, error) {
	return m.GetFn(key)
}

func (m *MockStorage) Set(key string, node Node) {
	m.SetFn(key, node)
}

func (m *MockStorage) Del(key string) int {
	return m.DelFn(key)
}
