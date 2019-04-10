package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
)

type Type byte

const (
	TypeSimpleString Type = '+'
	TypeError        Type = '-'
	TypeInt          Type = ':'
	TypeBulkString   Type = '$'
	TypeArray        Type = '*'
)

var (
	crlf = []byte{'\r', '\n'}
)

type Message interface {
	Type() Type

	marshal() ([]byte, error)
	unmarshal(*bufio.Reader) error
}

func MarshalMessage(m Message) ([]byte, error) {
	return m.marshal()
}

func ParseMessage(b []byte) (Message, error) {
	buf := bufio.NewReader(bytes.NewReader(b))

	return ReadMessage(buf)
}

func ReadMessage(buf *bufio.Reader) (Message, error) {
	typ, _ := buf.ReadByte()
	var m Message
	switch t := Type(typ); t {
	case TypeSimpleString:
		m = new(SimpleString)
	case TypeError:
		m = new(Error)
	case TypeInt:
		m = new(Int)
	case TypeBulkString:
		m = new(BulkString)
	case TypeArray:
		m = new(Array)
	default:
		return nil, fmt.Errorf("unrecognized type: %s", string(t))
	}

	if err := m.unmarshal(buf); err != nil {
		return nil, err
	}

	return m, nil
}

type SimpleString struct {
	Value string
}

var _ Message = &SimpleString{}

func (ss *SimpleString) Type() Type { return TypeSimpleString }

func (ss *SimpleString) marshal() ([]byte, error) {
	// TODO: probably don't need a bytes buffer here
	var buf bytes.Buffer
	buf.WriteByte(byte(ss.Type()))
	buf.WriteString(ss.Value)
	buf.Write(crlf)

	return buf.Bytes(), nil
}

func (ss *SimpleString) unmarshal(buf *bufio.Reader) error {
	b, err := buf.ReadBytes('\n')
	if err != nil {
		return err
	}

	// remove CRLF
	ss.Value = string(b[:len(b)-2])
	return nil
}

type Error struct {
	Value string
}

var _ Message = &Error{}

func (e *Error) Type() Type { return TypeError }

func (e *Error) marshal() ([]byte, error) {
	// TODO: need bytes.Buffer?
	var buf bytes.Buffer
	buf.WriteByte(byte(e.Type()))
	buf.WriteString(e.Value)
	buf.Write(crlf)

	return buf.Bytes(), nil
}

func (e *Error) unmarshal(buf *bufio.Reader) error {
	b, err := buf.ReadBytes('\n')
	if err != nil {
		return err
	}

	// remove CRLF
	e.Value = string(b[:len(b)-2])
	return nil
}

type Int struct {
	Value int64
}

var _ Message = &Int{}

func (i *Int) Type() Type { return TypeInt }

func (i *Int) marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(i.Type()))
	buf.WriteString(strconv.FormatInt(i.Value, 10))
	buf.Write(crlf)

	return buf.Bytes(), nil
}

func (i *Int) unmarshal(buf *bufio.Reader) error {
	i.Value, _ = readInt(buf)

	return nil
}

type BulkString struct {
	Value []byte
}

var _ Message = &BulkString{}

func (b *BulkString) Type() Type { return TypeBulkString }

func (b *BulkString) marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(b.Type()))

	if b.Value == nil {
		buf.WriteString("-1")
	} else {
		buf.WriteString(strconv.Itoa(len(b.Value)))
		buf.Write(crlf)
		buf.Write(b.Value)
	}
	buf.Write(crlf)

	return buf.Bytes(), nil
}

func (b *BulkString) unmarshal(buf *bufio.Reader) error {
	bLen, err := readInt(buf)
	if err != nil {
		return err
	}

	if bLen < 0 { // null bulk string
		b.Value = nil
		return nil
	}

	bulk, err := readNext(buf, bLen)
	if err != nil {
		return err
	}

	// discard CRLF
	buf.Discard(2)

	b.Value = bulk

	return nil
}

type Array struct {
	Value []Message
}

var _ Message = &Array{}

func (a *Array) Type() Type { return TypeArray }

func (a *Array) marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(a.Type()))

	if a.Value == nil {
		buf.WriteString("-1")
		buf.Write(crlf)
	} else {
		buf.WriteString(strconv.Itoa(len(a.Value)))
		buf.Write(crlf)

		for _, m := range a.Value {
			b, _ := m.marshal()
			buf.Write(b)
		}
	}

	return buf.Bytes(), nil
}

func (a *Array) unmarshal(buf *bufio.Reader) error {
	aLen, err := readInt(buf)
	if err != nil {
		return err
	}

	if aLen < 0 { // null array
		a.Value = nil
		return nil
	}

	a.Value = make([]Message, aLen)

	var i int64
	for i = 0; i < aLen; i++ {
		msg, err := ReadMessage(buf)
		if err != nil {
			return err
		}

		a.Value[i] = msg
	}

	return nil
}

func readInt(buf *bufio.Reader) (int64, error) {
	var n int64

	negative := false

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}

		if b == '-' {
			negative = true
			continue
		}

		// discard CRLF
		if b == '\r' {
			buf.Discard(1)
			break
		}

		n = (n * 10) + int64(rune(b)-'0')
	}

	if negative {
		n = -(n)
	}

	return n, nil
}

func readNext(buf *bufio.Reader, bLen int64) ([]byte, error) {
	bbuf := make([]byte, bLen)
	var i int64
	for i = 0; i < bLen; i++ {
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		bbuf[i] = b
	}
	return bbuf, nil
}
