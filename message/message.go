package message

type MessageType int

const (
	Text MessageType = iota
	Login
	Logout
	SingleTick
)

type Message struct {
	Type MessageType
	Data string
	To   string
	From string
}
