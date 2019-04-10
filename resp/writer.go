package resp

import (
	"io"
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

func (w *Writer) WriteMessage(msg Message) error {
	buf, err := msg.marshal()
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}
