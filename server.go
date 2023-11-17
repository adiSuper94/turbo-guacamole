package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

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

type Server struct {
	clients map[string]*websocket.Conn
}

func (s Server) HandleMessage(msg Message, ctx context.Context, clientAddr string, clientConn *websocket.Conn) error {
	switch msg.Type {
	case Text:
		clientConn, ok := s.clients[msg.To]
		if !ok {
			fmt.Fprintln(os.Stderr, "Error4: `to` is offline")
			break
		}
		msg.From, msg.To = msg.To, msg.From
		bytes, err := json.Marshal(msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error5: ", err)
			return errors.New("Could not marshal message")
		}
		clientConn.Write(ctx, websocket.MessageText, bytes)
	case Logout:
		delete(s.clients, clientAddr)
	case Login:
		s.clients[clientAddr] = clientConn
	}
	return nil
}

func main() {
	server := Server{
		clients: make(map[string]*websocket.Conn),
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
		server.clients[addr] = c
		defer delete(server.clients, addr)
		for {
			_, rawMessageBytes, err := c.Read(ctx)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error2: ", err)
				break
			}
			var msg Message
			err = json.Unmarshal(rawMessageBytes, &msg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error3: ", err)
				break
			}
			msg.From = addr
			log.Printf("Received: %s", msg.Data)
			err = server.HandleMessage(msg, ctx, addr, c)
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
