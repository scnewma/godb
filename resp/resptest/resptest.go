package resptest

import (
	"strings"

	"github.com/scnewma/godb/resp"
)

func NewRequest(command string) *resp.Request {
	return &resp.Request{
		RawMessage: asServerCommand(command),
	}
}

func NewRequestMessage(msg *resp.Array) *resp.Request {
	return &resp.Request{
		RawMessage: msg,
	}
}

func asServerCommand(s string) *resp.Array {
	parts := strings.Split(s, " ")

	var msgs []resp.Message
	for _, p := range parts {
		msgs = append(msgs, &resp.BulkString{[]byte(p)})
	}

	return &resp.Array{msgs}
}
