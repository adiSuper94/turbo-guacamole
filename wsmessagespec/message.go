package wsmessagespec

import "github.com/google/uuid"

type WSMessageType int

const (
	Text WSMessageType = iota
	Login
	LoginAck
	Logout
	SingleTick
	AddMemberToChatRoom
)

type WSMessage struct {
	Id   uuid.UUID // message id
	Type WSMessageType
	Data string
	To   uuid.UUID //chatroom id
	From string    // self username
}
