package main

import (
	"bufio"
	"errors"
)

const (
	TypeSimpleString byte = '+'
	TypeError        byte = '-'
	TypeInt          byte = ':'
	TypeBulkString   byte = '$'
	TypeArray        byte = '*'
)

type Message struct {
	Type   byte
	String string
	Error  string
	Int    int64
	Bulk   []byte
	Array  []*Message
}

type Parser struct {
	r *bufio.Reader
}

func NewParser(r *bufio.Reader) *Parser {
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

	bulk, err := p.readToLF()
	if err != nil {
		return nil, err
	}

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
	for i := 0; i < int(aLen); i++ {
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
	line, err := p.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line[:len(line)-2], err
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
			p.r.Discard(1) // discard \n
			break
		}

		i = (i * 10) + int64(rune(b)-'0')
	}

	if negative {
		i = -(i)
	}

	return i, nil
}
