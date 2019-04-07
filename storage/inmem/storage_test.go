package inmem

import (
	"testing"

	"github.com/scnewma/godb/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSet(t *testing.T) {
	db := NewStorage()

	db.Set("test", storage.NewNode("value"))

	n, err := db.Get("test")
	require.NoError(t, err)
	assert.Equal(t, "value", n.Value())
}

func TestGetDoesNotExist(t *testing.T) {
	db := NewStorage()

	_, err := db.Get("test")
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestDel(t *testing.T) {
	require := require.New(t)

	db := NewStorage()

	db.Set("test", storage.NewNode("value"))

	_, err := db.Get("test")
	require.NoError(err)

	require.Equal(1, db.Del("test"))

	_, err = db.Get("test")
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestDelDoesNotExist(t *testing.T) {
	db := NewStorage()

	assert.Equal(t, 0, db.Del("test"))
}
