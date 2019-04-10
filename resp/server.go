package resp

import (
	"bufio"
	"errors"
	"net"
	"time"
)

const (
	idleTimeout = 60 * time.Second
)

type conn struct {
	server *server

	rwc net.Conn
}

func (c *conn) serve() {
	defer c.rwc.Close()
	c.rwc.SetReadDeadline(time.Now().Add(idleTimeout))

	bufr := bufio.NewReader(c.rwc)
	bufw := bufio.NewWriter(c.rwc)

	respw := NewWriter(bufw)

	for {
		msg, err := ReadMessage(bufr)
		if err != nil {
			respw.WriteMessage(&Error{err.Error()})

			// this will close the connection if the read deadline
			// is exceeded or the client passes in an unparseable
			// message.
			break
		}

		arr, ok := msg.(*Array)
		if !ok {
			respw.WriteMessage(&Error{"invalid command"})
		}

		c.server.Handler.Serve(respw, &Request{
			RawMessage: arr,
		})

		bufw.Flush()

		c.rwc.SetReadDeadline(time.Now().Add(idleTimeout))
	}
}

type Request struct {
	RawMessage *Array

	command string
	args    [][]byte
}

func (r *Request) ParseCommand() error {
	for i, a := range r.RawMessage.Value {
		bs, ok := a.(*BulkString)
		if !ok {
			return errors.New("invalid command")
		}

		if i == 0 {
			r.command = string(bs.Value)
			continue
		}

		r.args = append(r.args, bs.Value)
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
	WriteMessage(Message) error
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
	for {
		rw, err := ln.Accept()
		if err != nil {
			return err
		}

		c := srv.newConn(rw)
		go c.serve()
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
