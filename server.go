package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
)

func NewServer(addr string) (*TCPServer, error) {
	return &TCPServer{
		addr: addr,
	}, nil
}

type TCPServer struct {
	addr   string
	server net.Listener
}

func (s *TCPServer) Run(db Database) error {
	var err error
	s.server, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return s.handleConnections(db)
}

func (s *TCPServer) Close() error {
	return s.server.Close()
}

func (s *TCPServer) handleConnections(db Database) (err error) {
	fmt.Printf("handling connections\n")

	for {
		fmt.Printf("waiting for new connection\n")
		conn, err := s.server.Accept()
		if err != nil || conn == nil {
			err = errors.New("could not accept connection")
			break
		}

		fmt.Printf("new connection from %s\n", conn.RemoteAddr().String())
		go s.handleConnection(conn, db)
	}
	return
}

func (s *TCPServer) handleConnection(conn net.Conn, db Database) {
	defer conn.Close()

	fmt.Printf("handling connection from %s\n", conn.RemoteAddr().String())

	r := bufio.NewReader(conn)
	rw := bufio.NewReadWriter(r, bufio.NewWriter(conn))
	parser := NewParser(r)
	writer := NewWriter(rw)
	for {
		msg, err := parser.Parse()
		if err != nil {
			writer.WriteMessage(NewErrorMessage("something went wrong"))
			rw.Flush()
			return
		}

		cmd, err := ParseCommand(msg)
		if err != nil {
			writer.WriteMessage(NewErrorMessage("unknown command"))
			rw.Flush()
			return
		}

		resp := cmd.Execute(db)
		writer.WriteMessage(resp)
		rw.Flush()
	}
}
