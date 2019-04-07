package executor

import (
	"github.com/scnewma/godb/resp"
)

func NewHandler(executor Executor) *handler {
	return &handler{
		executor: executor,
	}
}

type handler struct {
	executor Executor
}

func (h *handler) Serve(w resp.ResponseWriter, r *resp.Request) {
	response := h.executor.Execute(Command{
		Name: r.Command(),
		Args: r.Args(),
	})
	w.WriteMessage(response)
}
