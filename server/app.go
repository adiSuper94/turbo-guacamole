package main

import (
	"adisuper94/turboguac/server/generated"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"adisuper94/turboguac/wsmessagespec"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

func getChatRooms(ctx context.Context, username string) ([]generated.ChatRoom, error) {
	queries := GetQueries()
	chatRooms, err := queries.GetChatRoomDetailsByUsername(ctx, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting chatrooms\n%v\n", err)
		return nil, err
	}
	return chatRooms, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func main() {
	conn := getDBConn()
	defer conn.Close()
	server := Server{
		clients: make(map[string]ActiveClient),
	}

	router := http.NewServeMux()
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error1: ", err)
			return
		}
		defer c.CloseNow()
		addr := r.RemoteAddr
		log.Printf("Client connected: %s", addr)
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Minute)
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
			if client != nil {
				defer delete(server.clients, client.userName)
			}
			if err != nil {
				break
			}
		}

		c.Close(websocket.StatusNormalClosure, "")
	})

	router.HandleFunc("GET /online-users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Online Users")
		var activeUsers []string
		for k := range server.clients {
			activeUsers = append(activeUsers, k)
		}
		json.NewEncoder(w).Encode(activeUsers)
	})

	router.HandleFunc("GET /chatrooms", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Chatrooms")
		chatRooms, err := getChatRooms(r.Context(), r.FormValue("username"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(chatRooms)

	})

	router.HandleFunc("POST /chatrooms", func(w http.ResponseWriter, r *http.Request) {
		chatRoomName := r.FormValue("chatroom_name")
		username := r.FormValue("username")
		chatRoom, err := createChatRoom(r.Context(), username, chatRoomName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(chatRoom)
	})

	router.HandleFunc("GET /messages", func(w http.ResponseWriter, r *http.Request) {
		queries := GetQueries()
		chatRoomIdStr := r.FormValue("chatRoomId")
		chatRoomId, err := uuid.Parse(chatRoomIdStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while parsing chatroomid\n%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messages, err := queries.GetMessagesByChatRoomId(r.Context(), chatRoomId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting messages\n%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(messages)
	})

	router.HandleFunc("GET /dms", func(w http.ResponseWriter, r *http.Request) {
		queries := GetQueries()
		username := r.FormValue("username")
		dms, err := queries.GetDMs(r.Context(), username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting dms\n%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(dms)
	})
  layeredRouter := corsMiddleware(router)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: layeredRouter,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func askQuestion(question string) bool {
	fmt.Printf("%s (y/N)?", question)
	var isHttps bool
	stdinReader := bufio.NewReader(os.Stdin)
Loop:
	for {
		byte, _ := stdinReader.ReadByte()
		switch byte {
		case 'y', 'Y':
			isHttps = true
			break Loop
		case 'n', 'N':
			isHttps = false
			break Loop
		case '\n', '\r', ' ':
			continue
		default:
			fmt.Println("Please enter y or n")
		}
	}
	return isHttps
}
