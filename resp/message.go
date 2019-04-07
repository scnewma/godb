package resp

const (
	TypeSimpleString byte = '+'
	TypeError        byte = '-'
	TypeInt          byte = ':'
	TypeBulkString   byte = '$'
	TypeArray        byte = '*'
)

type Message struct {
	Type   byte
	String string
	Error  string
	Int    int64
	Bulk   []byte
	Array  []*Message
}

func NewSimpleStringMessage(s string) *Message {
	return &Message{
		Type:   TypeSimpleString,
		String: s,
	}
}

func NewErrorMessage(e string) *Message {
	return &Message{
		Type:  TypeError,
		Error: e,
	}
}

func NewIntMessage(i int64) *Message {
	return &Message{
		Type: TypeInt,
		Int:  i,
	}
}

func NewBulkStringMessage(buf []byte) *Message {
	return &Message{
		Type: TypeBulkString,
		Bulk: buf,
	}
}

func NewNilBulkStringMessage() *Message {
	return &Message{
		Type: TypeBulkString,
		Bulk: nil,
	}
}

func NewArrayMessage(msgs ...*Message) *Message {
	if msgs == nil {
		msgs = []*Message{}
	}
	return &Message{
		Type:  TypeArray,
		Array: msgs,
	}
}

func NewNilArrayMessage() *Message {
	return &Message{
		Type:  TypeArray,
		Array: nil,
	}
}
