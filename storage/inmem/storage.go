package inmem

import (
	"sync"

	"github.com/scnewma/godb/storage"
)

func NewStorage() *database {
	return &database{data: make(map[string]*node, 1024)}
}

type database struct {
	sync.RWMutex

	data map[string]*node
}

func (db *database) Get(key string) (storage.Node, error) {
	db.RLock()

	n, ok := db.data[key]
	db.RUnlock()

	if !ok {
		return nil, storage.ErrKeyNotFound
	}

	return n, nil
}

func (db *database) Set(key string, n storage.Node) {
	db.Lock()
	db.data[key] = newNode(n)
	db.Unlock()
}

func (db *database) Del(key string) int {
	db.Lock()
	defer db.Unlock()

	_, ok := db.data[key]
	if ok {
		delete(db.data, key)
		return 1
	}

	return 0
}

type node struct {
	sync.RWMutex

	storage.Node
}

func newNode(n storage.Node) *node {
	return &node{Node: n}
}
