package message

type MessageType int

const (
	Text MessageType = iota
	Login
	LoginAck
	Logout
	SingleTick
)

type Message struct {
	Type MessageType
	Data string
	To   string
	From string
}
