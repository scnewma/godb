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

func NewRequestMessage(msg *resp.Message) *resp.Request {
	return &resp.Request{
		RawMessage: msg,
	}
}

func asServerCommand(s string) *resp.Message {
	parts := strings.Split(s, " ")

	var msgs []*resp.Message
	for _, p := range parts {
		msgs = append(msgs, resp.NewBulkStringMessage([]byte(p)))
	}

	return resp.NewArrayMessage(msgs...)
}
