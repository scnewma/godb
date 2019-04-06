package main

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

const (
	TypeSimpleString byte = '+'
	TypeError        byte = '-'
	TypeInt          byte = ':'
	TypeBulkString   byte = '$'
	TypeArray        byte = '*'

	CR byte = '\r'
	LF byte = '\n'
)

type Message struct {
	Type   byte
	String string
	Error  string
	Int    int64
	Bulk   []byte
	Array  []*Message
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *Writer) WriteMessage(msg *Message) error {
	var b bytes.Buffer
	w.write(msg, &b)

	_, err := b.WriteTo(w)
	return err
}

func (w *Writer) write(msg *Message, b *bytes.Buffer) {
	b.WriteByte(msg.Type)
	switch msg.Type {
	case TypeSimpleString:
		b.WriteString(msg.String)
		b.WriteByte(CR)
		b.WriteByte(LF)
	case TypeError:
		b.WriteString(msg.Error)
		b.WriteByte(CR)
		b.WriteByte(LF)
	case TypeInt:
		b.WriteString(strconv.FormatInt(msg.Int, 10))
		b.WriteByte(CR)
		b.WriteByte(LF)
	case TypeBulkString:
		if msg.Bulk == nil {
			b.WriteString("-1")
		} else {
			b.WriteString(strconv.Itoa(len(msg.Bulk)))
			b.WriteByte(CR)
			b.WriteByte(LF)
			b.Write(msg.Bulk)
		}
		b.WriteByte(CR)
		b.WriteByte(LF)
	case TypeArray:
		if msg.Array == nil {
			b.WriteString("-1")
			b.WriteByte(CR)
			b.WriteByte(LF)
		} else {
			b.WriteString(strconv.Itoa(len(msg.Array)))
			b.WriteByte(CR)
			b.WriteByte(LF)

			for _, m := range msg.Array {
				w.write(m, b)
			}
		}
	}
}

type Parser struct {
	r *bytes.Reader
}

func NewParser(r *bytes.Reader) *Parser {
	return &Parser{r: r}
}

func (p *Parser) Parse() (*Message, error) {
	b, err := p.r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case TypeSimpleString:
		return p.parseSimpleString()
	case TypeError:
		return p.parseError()
	case TypeInt:
		return p.parseInt()
	case TypeBulkString:
		return p.parseBulkString()
	case TypeArray:
		return p.parseArray()
	}

	return nil, errors.New("unable to parse")
}

func (p *Parser) parseSimpleString() (*Message, error) {
	line, err := p.readToLF()
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:   TypeSimpleString,
		String: string(line),
	}, nil
}

func (p *Parser) parseError() (*Message, error) {
	line, err := p.readToLF()
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:  TypeError,
		Error: string(line),
	}, nil
}

func (p *Parser) parseInt() (*Message, error) {
	i, err := p.readInt()
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: TypeInt,
		Int:  i,
	}, nil
}

func (p *Parser) parseBulkString() (*Message, error) {
	bLen, err := p.readInt()
	if err != nil {
		return nil, errors.New("invalid bulk string length")
	}

	if bLen < 0 { // null bulk string
		return &Message{
			Type: TypeBulkString,
			Bulk: nil,
		}, nil
	}

	bulk, err := p.readNext(bLen)
	if err != nil {
		return nil, err
	}

	// discard \r\n
	p.r.Seek(2, io.SeekCurrent)

	return &Message{
		Type: TypeBulkString,
		Bulk: bulk,
	}, nil
}

func (p *Parser) parseArray() (*Message, error) {
	aLen, err := p.readInt()
	if err != nil {
		return nil, err
	}

	if aLen < 0 {
		return &Message{
			Type:  TypeArray,
			Array: nil,
		}, nil
	}

	msgs := make([]*Message, aLen)
	var i int64
	for i = 0; i < aLen; i++ {
		msg, err := p.Parse()
		if err != nil {
			return nil, err
		}

		msgs[i] = msg
	}

	return &Message{
		Type:  TypeArray,
		Array: msgs,
	}, nil
}

func (p *Parser) readToLF() ([]byte, error) {
	buf := make([]byte, 4096)
	i := 0
	for {
		b, err := p.r.ReadByte()
		if err != nil {
			return nil, err
		}

		if b == '\r' {
			p.r.Seek(1, io.SeekCurrent) // discard \n
			break
		}

		buf[i] = b
		i++
	}

	return buf[:i], nil
}

func (p *Parser) readNext(bLen int64) ([]byte, error) {
	buf := make([]byte, bLen)
	var i int64
	for i = 0; i < bLen; i++ {
		b, err := p.r.ReadByte()
		if err != nil {
			return nil, err
		}

		buf[i] = b
	}
	return buf, nil
}

func (p *Parser) readInt() (int64, error) {
	var i int64

	negative := false

	for {
		b, err := p.r.ReadByte()
		if err != nil {
			return 0, err
		}

		if b == '-' {
			negative = true
			continue
		}

		if b == '\r' {
			p.r.Seek(1, io.SeekCurrent) // discard \n
			break
		}

		i = (i * 10) + int64(rune(b)-'0')
	}

	if negative {
		i = -(i)
	}

	return i, nil
}
