package main

import (
	"flag"
	"fmt"

	"github.com/scnewma/godb/executor"
	"github.com/scnewma/godb/resp"
	"github.com/scnewma/godb/storage/inmem"
)

func main() {
	addr := flag.String("addr", ":1123", "tcp listen addr")
	flag.Parse()

	db := inmem.NewStorage()
	exctr := executor.NewExecutor(db)
	handler := executor.NewHandler(exctr)

	fmt.Printf("Serving on %s\n", *addr)
	resp.ListenAndServe(*addr, handler)
}
