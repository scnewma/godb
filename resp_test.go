package main

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		name     string
		payload  string
		expected *Message
	}{
		{"Simple String", "+OK\r\n", &Message{Type: TypeSimpleString, String: "OK"}},
		{"Error", "-Error message\r\n", &Message{Type: TypeError, Error: "Error message"}},
		{"Int", ":1000\r\n", &Message{Type: TypeInt, Int: 1000}},
		{"Negative Int", ":-1000\r\n", &Message{Type: TypeInt, Int: -1000}},
		{"Bulk String", "$6\r\nfoobar\r\n", &Message{Type: TypeBulkString, Bulk: []byte("foobar")}},
		{"Empty Bulk String", "$0\r\n\r\n", &Message{Type: TypeBulkString, Bulk: []byte("")}},
		{"Null Bulk String", "$-1\r\n", &Message{Type: TypeBulkString, Bulk: nil}},
		{"Empty Array", "*0\r\n", &Message{Type: TypeArray, Array: []*Message{}}},
		{"Bulk String Array", "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", &Message{
			Type: TypeArray,
			Array: []*Message{
				{Type: TypeBulkString, Bulk: []byte("foo")},
				{Type: TypeBulkString, Bulk: []byte("bar")},
			},
		}},
		{"Int Array", "*3\r\n:1\r\n:2\r\n:3\r\n", &Message{
			Type: TypeArray,
			Array: []*Message{
				{Type: TypeInt, Int: 1},
				{Type: TypeInt, Int: 2},
				{Type: TypeInt, Int: 3},
			},
		}},
		{"Mixed Array", "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n", &Message{
			Type: TypeArray,
			Array: []*Message{
				{Type: TypeInt, Int: 1},
				{Type: TypeInt, Int: 2},
				{Type: TypeInt, Int: 3},
				{Type: TypeInt, Int: 4},
				{Type: TypeBulkString, Bulk: []byte("foobar")},
			},
		}},
		{"Null Array", "*-1\r\n", &Message{Type: TypeArray, Array: nil}},
		{"Array of arrays", "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", &Message{
			Type: TypeArray,
			Array: []*Message{
				&Message{
					Type: TypeArray,
					Array: []*Message{
						{Type: TypeInt, Int: 1},
						{Type: TypeInt, Int: 2},
						{Type: TypeInt, Int: 3},
					},
				},
				&Message{
					Type: TypeArray,
					Array: []*Message{
						{Type: TypeSimpleString, String: "Foo"},
						{Type: TypeError, Error: "Bar"},
					},
				},
			},
		}},
		{"Null elements in array", "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n", &Message{
			Type: TypeArray,
			Array: []*Message{
				{Type: TypeBulkString, Bulk: []byte("foo")},
				{Type: TypeBulkString, Bulk: nil},
				{Type: TypeBulkString, Bulk: []byte("bar")},
			},
		}},
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

func newReader(val string) *bufio.Reader {
	return bufio.NewReader(bytes.NewReader([]byte(val)))
}
