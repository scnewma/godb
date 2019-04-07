package executor

import (
	"errors"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/scnewma/godb/resp"
	"github.com/scnewma/godb/storage"
)

const (
	GET = "GET"
	SET = "SET"
	DEL = "DEL"
)

var genericErrorMessage = resp.NewErrorMessage("something went wrong")

type Command struct {
	Name string
	Args [][]byte
}

type Executor interface {
	Execute(command Command) *resp.Message
}

type compositeExecutor struct {
	executorLookup map[string]executorFunc

	db storage.Storage
}

func NewExecutor(db storage.Storage) *compositeExecutor {
	return &compositeExecutor{
		executorLookup: map[string]executorFunc{
			GET: executorFunc(executeGet),
			SET: executorFunc(executeSet),
			DEL: executorFunc(executeDel),
		},
		db: db,
	}
}

func (ce *compositeExecutor) Execute(command Command) *resp.Message {
	commandName := strings.ToUpper(command.Name)
	executorFunc, ok := ce.executorLookup[commandName]
	if !ok {
		return resp.NewErrorMessage("unknown command")
	}

	return executorFunc(command.Args, ce.db)
}

type executorFunc func(args [][]byte, db storage.Storage) *resp.Message

type argExtractor struct {
	args [][]byte

	err error
}

func newArgExtractor(args [][]byte) *argExtractor {
	return &argExtractor{args: args}
}

func (e *argExtractor) ExtractStringAt(idx int) string {
	return string(e.ExtractAt(idx))
}

func (e *argExtractor) ExtractAt(idx int) []byte {
	if idx >= len(e.args) {
		e.err = multierror.Append(e.err, errors.New("not enough arguments"))
		return nil
	}

	return e.args[idx]
}

func (e *argExtractor) Err() error {
	return e.err
}

func (e *argExtractor) Error() string {
	return e.err.Error()
}

func executeGet(args [][]byte, db storage.Storage) *resp.Message {
	ae := newArgExtractor(args)
	key := ae.ExtractStringAt(0)
	if ae.Err() != nil {
		return resp.NewErrorMessage(ae.Error())
	}

	node, err := db.Get(key)
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return resp.NewNilBulkStringMessage()
		}

		return genericErrorMessage
	}

	return resp.NewBulkStringMessage(node.Value().([]byte))
}

func executeSet(args [][]byte, db storage.Storage) *resp.Message {
	ae := newArgExtractor(args)
	key := ae.ExtractStringAt(0)
	val := ae.ExtractAt(1)

	if ae.Err() != nil {
		return resp.NewErrorMessage(ae.Error())
	}

	db.Set(key, storage.NewNode(val))

	return resp.NewSimpleStringMessage("OK")
}

func executeDel(args [][]byte, db storage.Storage) *resp.Message {
	ae := newArgExtractor(args)
	key := ae.ExtractStringAt(0)
	if ae.Err() != nil {
		return resp.NewErrorMessage(ae.Error())
	}

	delCount := db.Del(key)

	return resp.NewIntMessage(int64(delCount))
}
