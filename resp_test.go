package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMessage(t *testing.T) {
	var tests = []struct {
		name     string
		expected string
		given    *Message
	}{
		{"write simple string", "+OK\r\n", NewSimpleStringMessage("OK")},
		{"write error", "-Error message\r\n", NewErrorMessage("Error message")},
		{"write int", ":1000\r\n", NewIntMessage(1000)},
		{"write bulk string", "$6\r\nfoobar\r\n", NewBulkStringMessage([]byte("foobar"))},
		{"write empty bulk string", "$0\r\n\r\n", NewBulkStringMessage([]byte(""))},
		{"write null bulk string", "$-1\r\n", NewBulkStringMessage(nil)},
		{"write array", "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", NewArrayMessage(
			NewBulkStringMessage([]byte("foo")),
			NewBulkStringMessage([]byte("bar")),
		)},
		{"write empty array", "*0\r\n", NewArrayMessage()},
		{"write null array", "*-1\r\n", NewNilArrayMessage()},
		{"write array with null elements", "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n", NewArrayMessage(
			NewBulkStringMessage([]byte("foo")),
			NewNilBulkStringMessage(),
			NewBulkStringMessage([]byte("bar")),
		)},
		{"write array of arrays", "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", NewArrayMessage(
			NewArrayMessage(
				NewIntMessage(1),
				NewIntMessage(2),
				NewIntMessage(3),
			),
			NewArrayMessage(
				NewSimpleStringMessage("Foo"),
				NewErrorMessage("Bar"),
			),
		)},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			w := NewWriter(&b)
			err := w.WriteMessage(tt.given)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, b.String())
		})
	}
}

func TestParse(t *testing.T) {
	var tests = []struct {
		name     string
		payload  string
		expected *Message
	}{
		{"Simple String", "+OK\r\n", NewSimpleStringMessage("OK")},
		{"Error", "-Error message\r\n", NewErrorMessage("Error message")},
		{"Int", ":1000\r\n", NewIntMessage(1000)},
		{"Negative Int", ":-1000\r\n", NewIntMessage(-1000)},
		{"Bulk String", "$6\r\nfoobar\r\n", NewBulkStringMessage([]byte("foobar"))},
		{"Empty Bulk String", "$0\r\n\r\n", NewBulkStringMessage([]byte(""))},
		{"Null Bulk String", "$-1\r\n", NewNilBulkStringMessage()},
		{"Empty Array", "*0\r\n", NewArrayMessage()},
		{"Bulk String Array", "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", NewArrayMessage(
			NewBulkStringMessage([]byte("foo")),
			NewBulkStringMessage([]byte("bar")),
		)},
		{"Int Array", "*3\r\n:1\r\n:2\r\n:3\r\n", NewArrayMessage(
			NewIntMessage(1),
			NewIntMessage(2),
			NewIntMessage(3),
		)},
		{"Mixed Array", "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n", NewArrayMessage(
			NewIntMessage(1),
			NewIntMessage(2),
			NewIntMessage(3),
			NewIntMessage(4),
			NewBulkStringMessage([]byte("foobar")),
		)},
		{"Null Array", "*-1\r\n", NewNilArrayMessage()},
		{"Array of arrays", "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", NewArrayMessage(
			NewArrayMessage(
				NewIntMessage(1),
				NewIntMessage(2),
				NewIntMessage(3),
			),
			NewArrayMessage(
				NewSimpleStringMessage("Foo"),
				NewErrorMessage("Bar"),
			),
		)},
		{"Null elements in array", "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n", NewArrayMessage(
			NewBulkStringMessage([]byte("foo")),
			NewNilBulkStringMessage(),
			NewBulkStringMessage([]byte("bar")),
		)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(newReader(tt.payload))
			actual, err := parser.Parse()
			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

var result *Message

func BenchmarkParseSimpleString(b *testing.B) {
	benchmarkParse("+OK\r\n", b)
}

func BenchmarkParseError(b *testing.B) {
	benchmarkParse("-Error message\r\n", b)
}

func BenchmarkInt(b *testing.B) {
	benchmarkParse(":1000\r\n", b)
}

func BenchmarkBulkString(b *testing.B) {
	benchmarkParse("$6\r\nfoobar\r\n", b)
}

func BenchmarkBulkStringArray(b *testing.B) {
	benchmarkParse("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", b)
}

func benchmarkParse(in string, b *testing.B) {
	var r *Message

	for i := 0; i < b.N; i++ {
		parser := NewParser(newReader(in))

		r, _ = parser.Parse()
	}

	result = r
}

func newReader(val string) *bytes.Reader {
	return bytes.NewReader([]byte(val))
}
