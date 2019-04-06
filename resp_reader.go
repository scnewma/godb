package main

import (
	"bufio"
	"errors"
)

type Parser struct {
	r *bufio.Reader
}

func NewParser(r *bufio.Reader) *Parser {
	return &Parser{r}
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
	p.r.Discard(2)

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
	buf, err := p.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return buf[:len(buf)-2], nil
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
			p.r.Discard(1)
			break
		}

		i = (i * 10) + int64(rune(b)-'0')
	}

	if negative {
		i = -(i)
	}

	return i, nil
}
