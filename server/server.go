package main

import (
	"adisuper94/turboguac/server/generated"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"adisuper94/turboguac/wsmessagespec"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type ActiveClient struct {
	addr     string
	conn     *websocket.Conn
	userName string
}

type Server struct {
	clients map[string]ActiveClient
}

func (s Server) sendMessage(ctx context.Context, msg wsmessagespec.WSMessage) error {
	queries := GetQueries()
	chatRoomId := msg.To
	chatMembers, err := queries.GetChatRoomMembers(ctx, chatRoomId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get chat_room members\n%v\n", err)
		return err
	}
	var senderId uuid.UUID
	var sender ActiveClient
	for _, member := range chatMembers {
		if member.Username == msg.From {
			ok := false
			sender, ok = s.clients[member.Username]
			if !ok {
				return errors.New("you cannot send a message if you are not logged in")
			}
			senderId = member.UserID
			break
		}
	}
	if senderId == uuid.Nil {
		return errors.New("you cannot send a message to a chatroom you are not a member of")
	}
	insertMessageParams := generated.InsertMessageParams{
		ChatRoomID: chatRoomId,
		SenderID:   senderId,
		Body:       msg.Data,
	}
	if msg.Id != uuid.Nil {
		insertMessageParams.ID = msg.Id
	}
	insertedMessage, err := queries.InsertMessage(ctx, insertMessageParams)
	msgAck := wsmessagespec.WSMessage{
		Type: wsmessagespec.SingleTick,
		Data: fmt.Sprintf("%s", insertedMessage.ID),
		From: "God",
		To:   uuid.Nil,
	}
	msgAckBytes, err := json.Marshal(msgAck)
	if err != nil {
		return errors.New("Could not marshal message")
	}
	sender.conn.Write(ctx, websocket.MessageText, msgAckBytes)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while inserting message\n%v\n", err)
		return err
	}
	for _, member := range chatMembers {
		memberUserName := member.Username
		insertMessageDeliveryParams := generated.InsertMessageDeliveryParams{
			MessageID:   insertedMessage.ID,
			ChatRoomID:  chatRoomId,
			RecipientID: member.UserID,
		}
		if memberUserName == msg.From {
			continue
		}
		recipient, ok := s.clients[memberUserName]
		if ok {
			bytes, err := json.Marshal(msg)
			if err != nil {
				return errors.New("Could not marshal message")
			}
			recipient.conn.Write(context.Background(), websocket.MessageText, bytes)
			insertMessageDeliveryParams.Delivered = true
		} else {
			insertMessageDeliveryParams.Delivered = false
		}
		_, err = queries.InsertMessageDelivery(ctx, insertMessageDeliveryParams)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while inserting message delivery\n%v\n", err)
			return err
		}
	}
	return nil
}

func (s Server) loginOrRegister(ctx context.Context, msg wsmessagespec.WSMessage, clientAddr string, clientConn *websocket.Conn) (*ActiveClient, error) {
	queries := GetQueries()
	user, err := queries.GetUserByUsername(ctx, msg.From)
	if err != nil {
		if err == sql.ErrNoRows {
			user, err = queries.InsertUser(ctx, generated.InsertUserParams{
				Username: msg.From,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while inserting new user. login failed\n%v\n", err)
				return nil, err
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error while getting user\n%v\n", err)
			return nil, err
		}
	}
	client := ActiveClient{
		conn:     clientConn,
		addr:     clientAddr,
		userName: user.Username,
	}
	s.clients[client.userName] = client
	reply := wsmessagespec.WSMessage{
		Type: wsmessagespec.LoginAck,
		Data: "Login successful",
		From: "God",
		To:   uuid.Nil,
	}
	bytes, err := json.Marshal(reply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error5: ", err)
		return nil, errors.New("Could not marshal message")
	}
	clientConn.Write(ctx, websocket.MessageText, bytes)
	return &client, nil
}

func (s Server) logout(ctx context.Context, msg wsmessagespec.WSMessage) {
	userName := msg.From
	_, ok := s.clients[userName]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s is already offline", userName)
	}
	delete(s.clients, userName)
}

func (s Server) HandleMessage(msg wsmessagespec.WSMessage, ctx context.Context, clientAddr string, clientConn *websocket.Conn) (*ActiveClient, error) {
	switch msg.Type {
	case wsmessagespec.Text:
		s.sendMessage(ctx, msg)
	case wsmessagespec.Logout:
		s.logout(ctx, msg)
	case wsmessagespec.Login:
		return s.loginOrRegister(ctx, msg, clientAddr, clientConn)
	}
	return nil, nil
}

func main() {
	conn := getDBConn()
	defer conn.Close()
	server := Server{
		clients: make(map[string]ActiveClient),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error1: ", err)
			return
		}
		defer c.CloseNow()
		addr := r.RemoteAddr
		log.Printf("Client connected: %s", addr)
		ctx, cancel := context.WithTimeout(r.Context(), 360*time.Second)
		defer cancel()
		defer delete(server.clients, addr)
		for {
			_, rawMessageBytes, err := c.Read(ctx)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error2: ", err)
				break
			}
			var msg wsmessagespec.WSMessage
			err = json.Unmarshal(rawMessageBytes, &msg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error3: ", err)
				break
			}
			log.Printf("Received: %s", msg.Data)
			client, err := server.HandleMessage(msg, ctx, addr, c)
			server.clients[client.userName] = *client
			defer delete(server.clients, client.userName)
			if err != nil {
				break
			}
		}

		c.Close(websocket.StatusNormalClosure, "")
	})

	http.HandleFunc("/activeUsers", func(w http.ResponseWriter, r *http.Request) {
		var activeUsers []string
		for k := range server.clients {
			activeUsers = append(activeUsers, k)
		}
		json.NewEncoder(w).Encode(activeUsers)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
