package main

import (
	"bytes"
	"errors"
	"io"
)

const defaultBufferSize = 1024

type Parser struct {
	r          *bytes.Reader
	bufferSize int64
}

func NewParser(r *bytes.Reader) *Parser {
	return NewParserSize(r, defaultBufferSize)
}

// NewParserSize will create a parser that will
// attempt to parse SimpleStrings and Errors with the
// provided buffer size. If the messages are longer than
// the provided size the buffer will be increased (so it does
// not cause out of bounds). You might want to adjust this
// size to decrease memory allocs; increasing performance.
func NewParserSize(r *bytes.Reader, size int64) *Parser {
	return &Parser{
		r:          r,
		bufferSize: size,
	}
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

	return NewSimpleStringMessage(string(line)), nil
}

func (p *Parser) parseError() (*Message, error) {
	line, err := p.readToLF()
	if err != nil {
		return nil, err
	}

	return NewErrorMessage(string(line)), nil
}

func (p *Parser) parseInt() (*Message, error) {
	i, err := p.readInt()
	if err != nil {
		return nil, err
	}

	return NewIntMessage(i), nil
}

func (p *Parser) parseBulkString() (*Message, error) {
	bLen, err := p.readInt()
	if err != nil {
		return nil, errors.New("invalid bulk string length")
	}

	if bLen < 0 { // null bulk string
		return NewBulkStringMessage(nil), nil
	}

	bulk, err := p.readNext(bLen)
	if err != nil {
		return nil, err
	}

	// discard \r\n
	p.r.Seek(2, io.SeekCurrent)

	return NewBulkStringMessage(bulk), nil
}

func (p *Parser) parseArray() (*Message, error) {
	aLen, err := p.readInt()
	if err != nil {
		return nil, err
	}

	if aLen < 0 {
		return NewNilArrayMessage(), nil
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

	return NewArrayMessage(msgs...), nil
}

func (p *Parser) readToLF() ([]byte, error) {
	buf := make([]byte, p.bufferSize)
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

		// this is a hack to prevent
		// a slice index out of range
		if i >= len(buf) {
			buf = append(buf, b)
		} else {
			buf[i] = b
		}
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
