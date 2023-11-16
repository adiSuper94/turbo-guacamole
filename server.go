package main

import (
	"context"
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
	CloseConnection
)

type Message struct {
	Type MessageType
	Data []byte
}

type Server struct {
}

func main() {
	clients := make(map[string]*websocket.Conn)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error1: ", err)
		}
		defer c.CloseNow()
		addr := r.RemoteAddr
		log.Printf("Client connected: %s", addr)
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		clients[addr] = c
		for {
			_, v, err := c.Read(ctx)
			if err != nil {
				delete(clients, addr)
				fmt.Fprintln(os.Stderr, "Error2: ", err)
				break
			}
			log.Printf("Received: %s", v)
		}

		c.Close(websocket.StatusNormalClosure, "")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
