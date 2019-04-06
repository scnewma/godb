package main

import "sync"

type Database interface {
	Get(key string) (*Node, error)
	Set(key string, node *Node)
	Del(key string) int
}

type database struct {
	sync.RWMutex

	data map[string]*Node
}

func NewDB() *database {
	return &database{data: make(map[string]*Node)}
}

func (db *database) Get(key string) (*Node, error) {
	db.RLock()

	n, ok := db.data[key]
	db.RUnlock()

	if !ok {
		return nil, ErrKeyNotFound
	}

	return n, nil
}

func (db *database) Set(key string, n *Node) {
	db.Lock()
	db.data[key] = n
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

type Node struct {
	sync.RWMutex

	Value interface{}
}

func NewNode(val interface{}) *Node {
	return &Node{Value: val}
}
