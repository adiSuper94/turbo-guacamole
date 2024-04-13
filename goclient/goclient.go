package goclient

import (
	"adisuper94/turboguac/wsmessagespec"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type TurboGuacClient struct {
	ctx      context.Context
	conn     *websocket.Conn
	username string
}

func NewTurboGuacClient(ctx context.Context, username string) (*TurboGuacClient, error) {
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket.Dial: %v\n", err)
		os.Exit(1)
	}
	tgc := TurboGuacClient{
		ctx:      ctx,
		conn:     conn,
		username: username,
	}
	err = tgc.loginOrRegister()
	if err != nil {
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
		fmt.Fprintf(os.Stderr, "loginOrRegister() failed in go-client: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "Logout() failed in go-client: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "SendMessage() failed in go-client: %v\n", err)
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
	err = tgc.conn.Write(tgc.ctx, websocket.MessageText, bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "conn.Write() failed in go-client\n")
		return err
	}
	return nil
}
