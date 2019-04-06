package main

import (
	"errors"
	"strings"
)

const (
	GET = "GET"
	SET = "SET"
	DEL = "DEL"
)

type Command interface {
	Execute(db Database) *Message
}

type GetCommand struct {
	Key string
}

var (
	genericErrorMessage = NewErrorMessage("something went wrong")
)

func (c *GetCommand) Execute(db Database) *Message {
	node, err := db.Get(c.Key)
	if err != nil {
		if err == ErrKeyNotFound {
			return NewNilBulkStringMessage()
		}

		return genericErrorMessage
	}

	return NewBulkStringMessage([]byte(node.Value.(string)))
}

type SetCommand struct {
	Key   string
	Value string
}

func (c *SetCommand) Execute(db Database) *Message {
	db.Set(c.Key, NewNode(c.Value))

	return NewSimpleStringMessage("OK")
}

type DelCommand struct {
	Key string
}

func (c *DelCommand) Execute(db Database) *Message {
	n := db.Del(c.Key)

	return NewIntMessage(int64(n))
}

var (
	errInvalidMessage = errors.New("server commands must be bulk array")
	errWrongArgCount  = errors.New("wrong number of arguments")
	errNoCommand      = errors.New("no command provided")
)

func ParseCommand(msg *Message) (Command, error) {
	if msg.Type != TypeArray {
		return nil, errInvalidMessage
	}

	arr := msg.Array

	if arr == nil || len(arr) == 0 {
		return nil, errNoCommand
	}

	nameMsg := arr[0]
	if nameMsg.Type != TypeBulkString {
		return nil, errInvalidMessage
	}

	switch strings.ToUpper(string(nameMsg.Bulk)) {
	case GET:
		return parseGetCommand(arr[1:])
	case SET:
		return parseSetCommand(arr[1:])
	case DEL:
		return parseDelCommand(arr[1:])
	}

	return nil, errors.New("unknown command")
}

func parseGetCommand(args []*Message) (*GetCommand, error) {
	if len(args) != 1 {
		return nil, errWrongArgCount
	}

	key, err := parseKey(args[0])
	if err != nil {
		return nil, err
	}

	return &GetCommand{
		Key: key,
	}, nil
}

func parseSetCommand(args []*Message) (*SetCommand, error) {
	if len(args) != 2 {
		return nil, errWrongArgCount
	}

	key, err := parseKey(args[0])
	if err != nil {
		return nil, err
	}

	val, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return &SetCommand{
		Key:   key,
		Value: val,
	}, nil
}

func parseDelCommand(args []*Message) (*DelCommand, error) {
	if len(args) != 1 {
		return nil, errWrongArgCount
	}

	key, err := parseKey(args[0])
	if err != nil {
		return nil, err
	}

	return &DelCommand{
		Key: key,
	}, nil
}

func parseKey(msg *Message) (string, error) {
	key, err := parseValue(msg)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(key) == "" {
		return "", errors.New("no key provided")
	}
	return key, nil
}

func parseValue(msg *Message) (string, error) {
	if msg.Type != TypeBulkString {
		return "", errInvalidMessage
	}
	return string(msg.Bulk), nil
}
