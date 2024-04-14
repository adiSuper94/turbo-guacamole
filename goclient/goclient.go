package goclient

import (
	"adisuper94/turboguac/wsmessagespec"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
	Id   uuid.UUID
	Name string
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

func (tgc TurboGuacClient) CreateChatRoom() error {
	createChatRoomRequest := wsmessagespec.WSMessage{
		Id:   uuid.New(),
		Type: wsmessagespec.CreateChatRoom,
		Data: "CreateChatRoom",
		To:   uuid.Nil,
		From: tgc.username,
	}
	err := tgc.sendWSMessage(createChatRoomRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateChatRoom() failed in go-client:\n")
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
func (tgc TurboGuacClient) GetMyChatRooms() ([]ChatRoom, error) {
	url := fmt.Sprintf("http://%s/chatrooms?username=%s", tgc.serverAddr, tgc.username)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Get(%s) failed in go-client\n", url)
		return nil, err
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "io.ReadAll() failed in go-client\n")
		return nil, err
	}
	var chatRooms []ChatRoom
	err = json.Unmarshal(bytes, &chatRooms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json.Unmarshal() failed in go-client\n")
		return nil, err
	}
	return chatRooms, nil
}

func (tgc TurboGuacClient) GetOnlineUsers() ([]string, error) {
	url := fmt.Sprintf("http://%s/online-users", tgc.serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "http.Get(%s) failed in go-client\n", url)
		return nil, err
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "io.ReadAll() failed in go-client\n")
		return nil, err
	}
	var onlineUsers []string
	err = json.Unmarshal(bytes, &onlineUsers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json.Unmarshal() failed in go-client\n")
		return nil, err
	}
	return onlineUsers, nil
}
