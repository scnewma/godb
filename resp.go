package main

import (
	"bufio"
	"errors"
	"strconv"
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
	line, err := p.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	line = line[:len(line)-2] // remove CRLF
	b := line[0]

	switch b {
	case TypeSimpleString:
		return &Message{
			Type:   TypeSimpleString,
			String: string(line[1:]),
		}, nil
	case TypeError:
		return &Message{
			Type:  TypeError,
			Error: string(line[1:]),
		}, nil
	case TypeInt:
		i, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, errors.New("invalid int")
		}

		return &Message{
			Type: TypeInt,
			Int:  i,
		}, nil
	case TypeBulkString:
		bLen, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, errors.New("invalid bulk string length")
		}

		if bLen < 0 { // null bulk string
			return &Message{
				Type: TypeBulkString,
				Bulk: nil,
			}, nil
		}

		bulk, err := p.r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		return &Message{
			Type: TypeBulkString,
			Bulk: bulk[:bLen],
		}, nil
	case TypeArray:
		aLen, err := strconv.ParseInt(string(line[1:]), 10, 64)
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

	return nil, errors.New("unable to parse")
}
