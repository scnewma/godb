package main

import (
	"fmt"
	"log"
)

func main() {
	addr := ":1123"
	srv, err := NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}

	db := NewDB()

	fmt.Printf("Serving on %s\n", addr)
	err = srv.Run(db)
	if err != nil {
		log.Println(err)
	}
	defer srv.Close()
}
