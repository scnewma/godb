package resp

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMessage(t *testing.T) {
	type expected struct {
		msg Message
		err error
	}
	var tests = []struct {
		name     string
		given    []byte
		expected expected
	}{
		{"SimpleString", []byte("+OK\r\n"), expected{&SimpleString{Value: "OK"}, nil}},
		{"Blank SimpleString", []byte("+\r\n"), expected{&SimpleString{Value: ""}, nil}},
		{"Error", []byte("-Error message\r\n"), expected{&Error{Value: "Error message"}, nil}},
		{"Blank Error", []byte("-\r\n"), expected{&Error{Value: ""}, nil}},
		{"Int", []byte(":1000\r\n"), expected{&Int{Value: 1000}, nil}},
		{"Bulk String", []byte("$6\r\nfoobar\r\n"), expected{&BulkString{Value: []byte("foobar")}, nil}},
		{"Empty Bulk String", []byte("$0\r\n\r\n"), expected{&BulkString{Value: []byte("")}, nil}},
		{"Null Bulk String", []byte("$-1\r\n"), expected{&BulkString{Value: nil}, nil}},
		{"Array", []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), expected{
			&Array{Value: []Message{
				&BulkString{Value: []byte("foo")},
				&BulkString{Value: []byte("bar")},
			}}, nil},
		},
		{"Empty Array", []byte("*0\r\n"), expected{
			&Array{Value: []Message{}}, nil},
		},
		{"Null Array", []byte("*-1\r\n"), expected{
			&Array{Value: nil}, nil},
		},
		{"Null Elements in Array", []byte("*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"), expected{
			&Array{Value: []Message{
				&BulkString{[]byte("foo")},
				&BulkString{},
				&BulkString{[]byte("bar")},
			}}, nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseMessage(tt.given)

			if tt.expected.msg != nil {
				assert.Equal(t, tt.expected.msg, actual)
			}

			if tt.expected.err != nil {
				assert.Equal(t, tt.expected.err, err)
			}
		})
	}
}

func TestMarshalMessage(t *testing.T) {
	type expected struct {
		msg []byte
		err error
	}
	var tests = []struct {
		name     string
		given    Message
		expected expected
	}{
		{"SimpleString", &SimpleString{"OK"}, expected{[]byte("+OK\r\n"), nil}},
		{"Blank SimpleString", &SimpleString{Value: ""}, expected{[]byte("+\r\n"), nil}},
		{"Error", &Error{"Error message"}, expected{[]byte("-Error message\r\n"), nil}},
		{"Blank Error", &Error{Value: ""}, expected{[]byte("-\r\n"), nil}},
		{"Int", &Int{Value: 1000}, expected{[]byte(":1000\r\n"), nil}},
		{"Bulk String", &BulkString{Value: []byte("foobar")}, expected{[]byte("$6\r\nfoobar\r\n"), nil}},
		{"Empty Bulk String", &BulkString{Value: []byte("")}, expected{[]byte("$0\r\n\r\n"), nil}},
		{"Null Bulk String", &BulkString{Value: nil}, expected{[]byte("$-1\r\n"), nil}},
		{"Array", &Array{Value: []Message{
			&BulkString{Value: []byte("foo")},
			&BulkString{Value: []byte("bar")},
		}}, expected{[]byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), nil},
		},
		{"Empty Array", &Array{Value: []Message{}}, expected{[]byte("*0\r\n"), nil}},
		{"Null Array", &Array{Value: nil}, expected{[]byte("*-1\r\n"), nil}},
		{"Null Elements in Array", &Array{Value: []Message{
			&BulkString{[]byte("foo")},
			&BulkString{},
			&BulkString{[]byte("bar")},
		}}, expected{[]byte("*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"), nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := MarshalMessage(tt.given)

			if tt.expected.msg != nil {
				assert.Equal(t, tt.expected.msg, actual)
			}

			if tt.expected.err != nil {
				assert.Equal(t, tt.expected.err, err)
			}
		})
	}
}

func BenchmarkParseSimpleString(b *testing.B) {
	benchmarkParseMessage("+OK\r\n", b)
}

func BenchmarkParseError(b *testing.B) {
	benchmarkParseMessage("-Error message\r\n", b)
}

func BenchmarkParseInt(b *testing.B) {
	benchmarkParseMessage(":1000\r\n", b)
}

func BenchmarkParseBulkString(b *testing.B) {
	benchmarkParseMessage("$6\r\nfoobar\r\n", b)
}

func BenchmarkParseArray(b *testing.B) {
	benchmarkParseMessage("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", b)
}

func BenchmarkMarshalSimpleString(b *testing.B) {
	benchmarkMarshalMessage(&SimpleString{"OK"}, b)
}

func BenchmarkMarshalError(b *testing.B) {
	benchmarkMarshalMessage(&Error{"Error message"}, b)
}

func BenchmarkMarshalInt(b *testing.B) {
	benchmarkMarshalMessage(&Int{1000}, b)
}

func BenchmarkMarshalBulkString(b *testing.B) {
	benchmarkMarshalMessage(&BulkString{[]byte("foobar")}, b)
}

func BenchmarkMarshalArray(b *testing.B) {
	benchmarkMarshalMessage(&Array{
		[]Message{
			&BulkString{[]byte("foo")},
			&BulkString{[]byte("bar")},
		},
	}, b)
}

var message Message
var bbuf []byte

func benchmarkParseMessage(msg string, b *testing.B) {
	var m Message

	bbuf := bytes.NewBuffer([]byte(msg))
	buf := bufio.NewReader(bbuf)

	for i := 0; i < b.N; i++ {
		buf.Reset(bbuf)

		m, _ = ReadMessage(buf)
	}

	message = m
}

func benchmarkMarshalMessage(msg Message, b *testing.B) {
	var bb []byte

	for i := 0; i < b.N; i++ {
		bb, _ = MarshalMessage(msg)
	}

	bbuf = bb
}
