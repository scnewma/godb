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
		ExecuteFn: func(command Command) resp.Message {
			got = command

			return &resp.SimpleString{"OK"}
		},
	}
	h := NewHandler(e)

	req := resptest.NewRequest("GET blah")
	rec := resptest.NewRecorder()
	h.Serve(rec, req)

	assert := assert.New(t)
	assert.Equal(Command{Name: "GET", Args: [][]byte{
		[]byte("blah"),
	}}, got)
	assert.Equal(1, rec.MessageCount())
	assert.Equal(&resp.SimpleString{"OK"}, rec.MessageAt(0))
}
