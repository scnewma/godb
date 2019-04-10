package resptest

import "github.com/scnewma/godb/resp"

type ResponseRecorder struct {
	Messages []resp.Message
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (rr *ResponseRecorder) WriteMessage(msg resp.Message) error {
	rr.Messages = append(rr.Messages, msg)
	return nil
}

// MessageAt returns the message that was recorded
// with the given index, where index=0 is the first
// mesage that was recorded.
//
// Returns nil if idx is greater than the number of recorded
// messages.
func (rr *ResponseRecorder) MessageAt(idx int) resp.Message {
	if idx >= len(rr.Messages) {
		return nil
	}

	return rr.Messages[idx]
}

func (rr *ResponseRecorder) MessageCount() int {
	return len(rr.Messages)
}
