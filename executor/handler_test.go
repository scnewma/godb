package executor

import (
	"testing"

	"github.com/scnewma/godb/resp"
	"github.com/scnewma/godb/resp/resptest"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	var got Command
	e := MockExecutor{
		ExecuteFn: func(command Command) *resp.Message {
			got = command

			return resp.NewSimpleStringMessage("OK")
		},
	}
	h := NewHandler(e)

	rec := resptest.NewRecorder()
	req := resptest.NewRequest("GET blah")
	h.Serve(rec, req)

	assert := assert.New(t)
	assert.Equal(Command{Name: "GET", Args: [][]byte{
		[]byte("blah"),
	}}, got)
	assert.Equal(1, rec.MessageCount())
	assert.Equal(resp.NewSimpleStringMessage("OK"), rec.MessageAt(0))
}
