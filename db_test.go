package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSet(t *testing.T) {
	db := NewDB()

	db.Set("test", NewNode("value"))

	n, err := db.Get("test")
	require.NoError(t, err)
	assert.Equal(t, "value", n.Value)
}

func TestGetDoesNotExist(t *testing.T) {
	db := NewDB()

	_, err := db.Get("test")
	assert.Equal(t, err, ErrKeyNotFound)
}

func TestDel(t *testing.T) {
	require := require.New(t)

	db := NewDB()

	db.Set("test", NewNode("value"))

	_, err := db.Get("test")
	require.NoError(err)

	require.Equal(1, db.Del("test"))

	_, err = db.Get("test")
	assert.Equal(t, err, ErrKeyNotFound)
}

func TestDelDoesNotExist(t *testing.T) {
	db := NewDB()

	assert.Equal(t, 0, db.Del("test"))
}
