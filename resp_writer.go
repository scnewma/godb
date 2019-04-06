package main

import (
	"bytes"
	"io"
	"strconv"
)

const (
	cr byte = '\r'
	lf byte = '\n'
)

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *Writer) WriteMessage(msg *Message) error {
	var b bytes.Buffer
	w.write(msg, &b)

	_, err := b.WriteTo(w)
	return err
}

func (w *Writer) write(msg *Message, b *bytes.Buffer) {
	b.WriteByte(msg.Type)
	switch msg.Type {
	case TypeSimpleString:
		b.WriteString(msg.String)
		b.WriteByte(cr)
		b.WriteByte(lf)
	case TypeError:
		b.WriteString(msg.Error)
		b.WriteByte(cr)
		b.WriteByte(lf)
	case TypeInt:
		b.WriteString(strconv.FormatInt(msg.Int, 10))
		b.WriteByte(cr)
		b.WriteByte(lf)
	case TypeBulkString:
		if msg.Bulk == nil {
			b.WriteString("-1")
		} else {
			b.WriteString(strconv.Itoa(len(msg.Bulk)))
			b.WriteByte(cr)
			b.WriteByte(lf)
			b.Write(msg.Bulk)
		}
		b.WriteByte(cr)
		b.WriteByte(lf)
	case TypeArray:
		arr := msg.Array
		if arr == nil {
			b.WriteString("-1")
			b.WriteByte(cr)
			b.WriteByte(lf)
		} else {
			b.WriteString(strconv.Itoa(len(arr)))
			b.WriteByte(cr)
			b.WriteByte(lf)

			for _, m := range arr {
				w.write(m, b)
			}
		}
	}
}
