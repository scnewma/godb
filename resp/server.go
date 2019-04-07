package resp

import (
	"bufio"
	"context"
	"errors"
	"net"
)

type conn struct {
	server *server

	rwc net.Conn

	bufr *bufio.Reader
	bufw *bufio.Writer
}

func (c *conn) serve(ctx context.Context) {
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c.bufr = bufio.NewReader(c.rwc)
	c.bufw = bufio.NewWriter(c.rwc)

	parser := NewParser(c.bufr)
	respw := NewWriter(c.bufw)

	for {
		msg, err := parser.Parse()
		if err != nil {
			respw.WriteMessage(NewErrorMessage(err.Error()))
			continue
		}

		c.server.Handler.Serve(respw, &Request{
			RawMessage: msg,
		})

		c.bufw.Flush()
	}
}

type Request struct {
	RawMessage *Message

	command string
	args    [][]byte
}

func (r *Request) ParseCommand() error {
	if r.RawMessage.Type != TypeArray {
		return errors.New("invalid command")
	}

	for i, msg := range r.RawMessage.Array {
		if msg.Type != TypeBulkString {
			return errors.New("invalid command")
		}

		if i == 0 {
			r.command = string(msg.Bulk)
			continue
		}

		r.args = append(r.args, msg.Bulk)
	}

	return nil
}

func (r *Request) Command() string {
	if r.command == "" {
		r.ParseCommand()
	}

	return r.command
}

func (r *Request) Args() [][]byte {
	if r.command == "" {
		r.ParseCommand()
	}

	return r.args
}

type ResponseWriter interface {
	WriteMessage(*Message) error
}

type HandlerFunc func(ResponseWriter, *Request)

func (hf HandlerFunc) Serve(w ResponseWriter, r *Request) {
	hf(w, r)
}

type Handler interface {
	Serve(ResponseWriter, *Request)
}

type server struct {
	Addr    string
	Handler Handler
}

func (srv *server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *server) Serve(ln net.Listener) error {
	ctx := context.Background()
	for {
		rw, err := ln.Accept()
		if err != nil {
			return err
		}

		c := srv.newConn(rw)
		go c.serve(ctx)
	}
}

func (srv *server) newConn(rwc net.Conn) *conn {
	return &conn{
		server: srv,
		rwc:    rwc,
	}
}

func ListenAndServe(addr string, handler Handler) error {
	server := &server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}
