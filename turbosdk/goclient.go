package turbosdk

import (
	"adisuper94/turboguac/wsmessagespec"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type TurboGuacClient struct {
	ctx        context.Context
	Conn       *websocket.Conn
	username   string
	serverAddr string
}

type ChatRoom struct {
	ID         uuid.UUID
	Name       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type DM struct {
	Username   string
	ChatRoomId uuid.UUID
}

func NewTurboGuacClient(ctx context.Context, username string, serverAddr string) (*TurboGuacClient, error) {
	conn, _, err := websocket.Dial(ctx, "ws://"+serverAddr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket.Dial: %v\n", err)
		os.Exit(1)
	}
	tgc := TurboGuacClient{
		ctx:        ctx,
		Conn:       conn,
		username:   username,
		serverAddr: serverAddr,
	}
	err = tgc.loginOrRegister()
	if err != nil {
		fmt.Fprintf(os.Stderr, "loginOrRegister() failed in go-client: \n")
		return nil, err
	}
	return &tgc, nil
}

func (tgc TurboGuacClient) GetUsername() string {
	return tgc.username
}

func (tgc TurboGuacClient) loginOrRegister() error {
	loginRequest := wsmessagespec.WSMessage{
		Id:   uuid.New(),
		Type: wsmessagespec.Login,
		Data: "Login",
		To:   uuid.Nil,
		From: tgc.username,
	}
	err := tgc.sendWSMessage(loginRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loginOrRegister() failed in go-client: \n")
		return err
	}
	return nil
}

type IncomingChat struct {
	Id      uuid.UUID // message id
	From    string    // username
	To      uuid.UUID // chatRoomId
	Message string
}

func (tgc TurboGuacClient) WSListen(channel chan IncomingChat) {
	for {
		msgType, msg, err := tgc.Conn.Read(tgc.ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "conn.Read() failed in go-client\n %v\n", err)
			return
		}
		if msgType != websocket.MessageText {
			fmt.Fprintf(os.Stderr, "msgType is not websocket.MessageText in go-client\n")
			continue
		}
		var wsMsg wsmessagespec.WSMessage
		err = json.Unmarshal(msg, &wsMsg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "json.Unmarshal() failed in go-client\n")
			continue
		}
		switch wsMsg.Type {
		case wsmessagespec.Text:
			incomingMsg := IncomingChat{
				Id:      wsMsg.Id,
				From:    wsMsg.From,
				To:      wsMsg.To,
				Message: wsMsg.Data,
			}
			channel <- incomingMsg
			// fmt.Println("Recieved Text")
		case wsmessagespec.LoginAck:
			// fmt.Println("Recieved Acknowledgement for Login")
		case wsmessagespec.SingleTick:
			// fmt.Println("Recieved SingleTick")
		}
	}
}

func (tgc TurboGuacClient) Logout() error {
	logoutRequest := wsmessagespec.WSMessage{
		Id:   uuid.New(),
		Type: wsmessagespec.Logout,
		Data: "Logout",
		To:   uuid.Nil,
		From: tgc.username,
	}
	err := tgc.sendWSMessage(logoutRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logout() failed in go-client: \n")
		return err
	}
	return nil
}

func (tgc TurboGuacClient) SendMessage(data string, toChatRoomId uuid.UUID) error {
	if toChatRoomId == uuid.Nil {
		// fmt.Fprintf(os.Stderr, "cannot call SendMessage() go-client if toChatRoomId is nil\n")
		return nil
	}
	message := wsmessagespec.WSMessage{
		Id:   uuid.New(),
		Type: wsmessagespec.Text,
		Data: data,
		To:   toChatRoomId,
		From: tgc.username,
	}
	err := tgc.sendWSMessage(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SendMessage() failed in go-client:\n")
		return err
	}
	return nil
}

func (tgc TurboGuacClient) AddMemberToChatRoom(chatRoomId uuid.UUID, memberUsername string) error {
	addMemberRequest := wsmessagespec.WSMessage{
		Id:   uuid.New(),
		Type: wsmessagespec.AddMemberToChatRoom,
		Data: memberUsername,
		To:   chatRoomId,
		From: tgc.username,
	}
	err := tgc.sendWSMessage(addMemberRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AddMemberToChatRoom() failed in go-client:\n")
		return err
	}
	return nil
}

func (tgc TurboGuacClient) sendWSMessage(msg wsmessagespec.WSMessage) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json.Marshal() failed in go-client\n")
		return err
	}
	err = tgc.Conn.Write(tgc.ctx, websocket.MessageText, bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "conn.Write() failed in go-client\n")
		return err
	}
	return nil
}

func (tgc TurboGuacClient) StartDM(username string) (uuid.UUID, error) {
	dms, err := tgc.GetDMs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetDMs() failed in go-client\n")
		return uuid.Nil, err
	}
	for _, dm := range dms {
		if dm.Username == username {
			return dm.ChatRoomId, nil
		}
	}
	chatRoom, err := tgc.CreateChatRoom()
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateChatRoom() failed in go-client\n")
		return uuid.Nil, err
	}
	err = tgc.AddMemberToChatRoom(chatRoom.ID, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AddMemberToChatRoom() failed in go-client\n")
		return uuid.Nil, err
	}
	return chatRoom.ID, nil
}

func (tgc TurboGuacClient) CreateChatRoom() (*ChatRoom, error) {
	url := fmt.Sprintf("http://%s/chatrooms?username=%s", tgc.serverAddr, tgc.username)
	return httpCall[*ChatRoom]("POST", url)
}

func (tgc TurboGuacClient) GetMyChatRooms() ([]ChatRoom, error) {
	url := fmt.Sprintf("http://%s/chatrooms?username=%s", tgc.serverAddr, tgc.username)
	return httpCall[[]ChatRoom]("GET", url)
}

func (tgc TurboGuacClient) GetOnlineUsers() ([]string, error) {
	url := fmt.Sprintf("http://%s/online-users", tgc.serverAddr)
	return httpCall[[]string]("GET", url)
}

func (tgc TurboGuacClient) GetDMs() ([]DM, error) {
	url := fmt.Sprintf("http://%s/dms?username=%s", tgc.serverAddr, tgc.username)
	return httpCall[[]DM]("GET", url)
}

func httpCall[T []string | []DM | []ChatRoom | *ChatRoom](method string, url string) (T, error) {
	var resp *http.Response
	var err error
	switch method {
	case "GET":
		resp, err = http.Get(url)
	case "POST":
		resp, err = http.Post(url, "application/json", nil)
	default:
		return nil, errors.New("Unsupported HTTP method: " + method)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Get(%s) failed in go-client\n", url)
		return nil, err
	}
	bytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "io.ReadAll() failed in go-client\n")
		return nil, err
	}
	var data T
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json.Unmarshal() failed in go-client\n")
		return nil, err
	}
	return data, nil
}
