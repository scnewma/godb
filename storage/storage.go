package storage

import "errors"

var ErrKeyNotFound = errors.New("key not found")

type Storage interface {
	Get(key string) (Node, error)
	Set(key string, node Node)
	Del(key string) int
}

type Node interface {
	Value() interface{}
}

type basicNode struct {
	value interface{}
}

func (n *basicNode) Value() interface{} {
	return n.value
}

func NewNode(val []byte) Node {
	return &basicNode{val}
}

func NewStringNode(val string) Node {
	return NewNode([]byte(val))
}
