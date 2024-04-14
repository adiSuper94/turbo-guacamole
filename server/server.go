package main

import (
	"adisuper94/turboguac/server/generated"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"adisuper94/turboguac/wsmessagespec"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while inserting message\n%v\n", err)
		return err
	}
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
		if err == pgx.ErrNoRows {
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

func (s Server) logout(msg wsmessagespec.WSMessage) {
	userName := msg.From
	_, ok := s.clients[userName]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s is already offline", userName)
	}
	delete(s.clients, userName)
}

func (s Server) createChatRoom(ctx context.Context, msg wsmessagespec.WSMessage) error {
	queries := GetQueries()
	chatRoom, err := queries.InsertChatRoom(ctx, generated.InsertChatRoomParams{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating chatroom\n%v\n", err)
		return err
	}
	user, err := queries.GetUserByUsername(ctx, msg.From)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting user\n%v\n", err)
		return err
	}
	_, err = queries.InsertMember(ctx, generated.InsertMemberParams{
		ChatRoomID: chatRoom.ID,
		UserID:     user.ID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while inserting chatroom member\n%v\n", err)
		return err
	}
	return nil
}

func (s Server) addMembertoChatRoom(ctx context.Context, msg wsmessagespec.WSMessage) error {
	queries := GetQueries()
	chatRoomId := msg.To
	user, err := queries.GetUserByUsername(ctx, msg.From)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting user. Please register before send requests\n%v\n", err)
		return err
	}
	chatRoomMembers, err := queries.GetChatRoomMembers(ctx, chatRoomId)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("No members in chatroom yet, (This should not happenn. But I'll allow this for now)")
			_, err := queries.GetChatRoomById(ctx, chatRoomId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting chatroom\n%v\n", err)
				return err
			}

		} else {
			fmt.Fprintf(os.Stderr, "Error while getting chatroom members\n%v\n", err)
			return err
		}
	}

	userInChatRoom := false
	for _, member := range chatRoomMembers {
		if member.Username == msg.Data {
			fmt.Println("User already in chatroom")
			return nil
		}
		if member.Username == user.Username {
			userInChatRoom = true
			continue
		}
	}
	if !userInChatRoom {
		fmt.Println("User not in chatroom")
		return errors.New("You cannot add a user to a chatroom you are not a member of")
	}
	userToAdd, err := queries.GetUserByUsername(ctx, msg.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting user to add\n%v\n", err)
		return err
	}
	_, err = queries.InsertMember(ctx, generated.InsertMemberParams{
		ChatRoomID: chatRoomId,
		UserID:     userToAdd.ID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while inserting chatroom member\n%v\n", err)
		return err
	}
	return nil
}

func (s Server) HandleMessage(msg wsmessagespec.WSMessage, ctx context.Context, clientAddr string, clientConn *websocket.Conn) (*ActiveClient, error) {
	switch msg.Type {
	case wsmessagespec.Text:
		return nil, s.sendMessage(ctx, msg)
	case wsmessagespec.Logout:
		s.logout(msg)
	case wsmessagespec.Login:
		return s.loginOrRegister(ctx, msg, clientAddr, clientConn)
	case wsmessagespec.CreateChatRoom:
		err := s.createChatRoom(ctx, msg)
		return nil, err
	case wsmessagespec.AddMemberToChatRoom:
		err := s.addMembertoChatRoom(ctx, msg)
		return nil, err
	case wsmessagespec.SingleTick, wsmessagespec.LoginAck:
		fmt.Println("Received a tick or login ack, ignoring")
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

	http.HandleFunc("/online-users", func(w http.ResponseWriter, r *http.Request) {
		var activeUsers []string
		for k := range server.clients {
			activeUsers = append(activeUsers, k)
		}
		json.NewEncoder(w).Encode(activeUsers)
	})

	http.HandleFunc("/chatrooms", func(w http.ResponseWriter, r *http.Request) {
		queries := GetQueries()
		username := r.FormValue("username")
		chatRooms, err := queries.GetChatRoomDetailsByUsername(r.Context(), username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting chatrooms\n%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(chatRooms)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
