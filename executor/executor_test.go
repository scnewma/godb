package executor

import (
	"strings"
	"testing"

	"github.com/scnewma/godb/resp"
	"github.com/scnewma/godb/storage"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	assert := assert.New(t)
	called := false
	db := &storage.MockStorage{
		GetFn: func(key string) (storage.Node, error) {
			called = true

			assert.Equal("blah", key)

			return storage.NewStringNode("value"), nil
		},
	}
	args := asArgs("blah")
	msg := executeGet(args, db)

	assert.True(called)
	assert.Equal(resp.NewBulkStringMessage([]byte("value")), msg)
}

func TestGetNotFound(t *testing.T) {
	assert := assert.New(t)
	called := false
	db := &storage.MockStorage{
		GetFn: func(key string) (storage.Node, error) {
			called = true

			assert.Equal("blah", key)

			return nil, storage.ErrKeyNotFound
		},
	}
	args := asArgs("blah")
	msg := executeGet(args, db)

	assert.True(called)
	assert.Equal(resp.NewNilBulkStringMessage(), msg)
}

func TestGetNoKey(t *testing.T) {
	assert := assert.New(t)
	db := &storage.MockStorage{
		GetFn: func(key string) (storage.Node, error) {
			t.Fatal("should not have been called")
			return nil, nil
		},
	}
	msg := executeGet([][]byte{}, db)

	assert.True(strings.Contains(msg.Error, "not enough arguments"))
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	called := false
	db := &storage.MockStorage{
		SetFn: func(key string, node storage.Node) {
			called = true

			assert.Equal("blah", key)
			assert.Equal([]byte("value"), node.Value())
		},
	}
	args := asArgs("blah", "value")
	msg := executeSet(args, db)

	assert.True(called)
	assert.Equal(resp.NewSimpleStringMessage("OK"), msg)
}

func TestSetNoKey(t *testing.T) {
	assert := assert.New(t)
	db := &storage.MockStorage{
		SetFn: func(key string, node storage.Node) {
			t.Fatal("should not have been called")
		},
	}
	msg := executeSet([][]byte{}, db)

	assert.True(strings.Contains(msg.Error, "not enough arguments"))
}

func TestSetNoVal(t *testing.T) {
	assert := assert.New(t)
	db := &storage.MockStorage{
		SetFn: func(key string, node storage.Node) {
			t.Fatal("should not have been called")
		},
	}
	msg := executeSet(asArgs("key"), db)

	assert.True(strings.Contains(msg.Error, "not enough arguments"))
}

func TestDel(t *testing.T) {
	assert := assert.New(t)
	called := false
	db := &storage.MockStorage{
		DelFn: func(key string) int {
			called = true

			assert.Equal("blah", key)

			return 1
		},
	}
	args := asArgs("blah")
	msg := executeDel(args, db)

	assert.True(called)
	assert.Equal(resp.NewIntMessage(1), msg)
}

func TestDelNoKey(t *testing.T) {
	assert := assert.New(t)
	db := &storage.MockStorage{
		DelFn: func(key string) int {
			t.Fatal("should not have been called")
			return 1
		},
	}
	msg := executeDel([][]byte{}, db)

	assert.True(strings.Contains(msg.Error, "not enough arguments"))
}

func asArgs(argStrs ...string) [][]byte {
	var args [][]byte
	for _, arg := range argStrs {
		args = append(args, []byte(arg))
	}
	return args
}
