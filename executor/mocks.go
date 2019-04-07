package executor

import "github.com/scnewma/godb/resp"

type MockExecutor struct {
	ExecuteFn func(Command) *resp.Message
}

func (e MockExecutor) Execute(command Command) *resp.Message {
	return e.ExecuteFn(command)
}
