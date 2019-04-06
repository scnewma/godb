package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func asMessage(s string) *Message {
	parts := strings.Split(s, " ")

	var msgs []*Message
	for _, p := range parts {
		msgs = append(msgs, NewBulkStringMessage([]byte(p)))
	}

	return NewArrayMessage(msgs...)
}

// sanity check the asMessage helper func
func TestAsMessage(t *testing.T) {
	assert := assert.New(t)

	msg := asMessage("GET blah")

	assert.Equal(TypeArray, msg.Type)
	assert.Len(msg.Array, 2)
	assert.Equal("GET", string(msg.Array[0].Bulk))
	assert.Equal("blah", string(msg.Array[1].Bulk))
}

func TestParseCommand(t *testing.T) {
	type expected struct {
		cmd Command
		err error
	}
	var tests = []struct {
		name     string
		given    *Message
		expected expected
	}{
		{
			"not array",
			NewIntMessage(100),
			expected{nil, errInvalidMessage},
		},
		{
			"not bulk string array",
			NewArrayMessage(
				NewIntMessage(100),
			),
			expected{nil, errInvalidMessage},
		},
		{
			"can parse GET command",
			asMessage("GET blah"),
			expected{&GetCommand{Key: "blah"}, nil},
		},
		{
			"GET command no key provided",
			asMessage("GET"),
			expected{nil, errWrongArgCount},
		},
		{
			"GET too many args",
			asMessage("GET arg1 arg2"),
			expected{nil, errWrongArgCount},
		},
		{
			"can parse SET command",
			asMessage("SET blah value"),
			expected{&SetCommand{Key: "blah", Value: "value"}, nil},
		},
		{
			"SET no args provided",
			asMessage("SET"),
			expected{nil, errWrongArgCount},
		},
		{
			"SET too many args",
			asMessage("SET key value other"),
			expected{nil, errWrongArgCount},
		},
		{
			"can parse DEL command",
			asMessage("DEL blah"),
			expected{&DelCommand{Key: "blah"}, nil},
		},
		{
			"DEL no key provided",
			asMessage("DEL"),
			expected{nil, errWrongArgCount},
		},
		{
			"DEL too many args",
			asMessage("DEL key other"),
			expected{nil, errWrongArgCount},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := ParseCommand(tt.given)

			// otherwise assert equal checks types of nil
			if tt.expected.cmd == nil {
				assert.Nil(actual)
			} else {
				assert.Equal(tt.expected.cmd, actual)
			}
			assert.Equal(tt.expected.err, err)
		})
	}
}
