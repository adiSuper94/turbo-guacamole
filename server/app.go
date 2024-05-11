package main

import (
	"adisuper94/turboguac/server/generated"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"adisuper94/turboguac/wsmessagespec"

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

var hbs *template.Template

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
func main() {
	hbs = template.Must(template.ParseGlob("*.html"))
	conn := getDBConn()
	defer conn.Close()
	server := Server{
		clients: make(map[string]ActiveClient),
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
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

	http.HandleFunc("/online-users", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
    fmt.Println("Online Users")
		var activeUsers []string
		for k := range server.clients {
			activeUsers = append(activeUsers, k)
		}
		if getAcceptHeader(r) == HTML {
			hbs.ExecuteTemplate(os.Stdin, "online-users", activeUsers)
			return
		}
		json.NewEncoder(w).Encode(activeUsers)
	})

	http.HandleFunc("/chatrooms", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		fmt.Println("Chatrooms")
		switch r.Method {
		case http.MethodGet:
			chatRooms, err := getChatRooms(r.Context(), r.FormValue("username"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if getAcceptHeader(r) == HTML {
				hbs.ExecuteTemplate(os.Stdin, "active-chat-rooms", chatRooms)
				return
			}
			json.NewEncoder(w).Encode(chatRooms)
		case http.MethodPost:
			chatRoomName := r.FormValue("chatroom_name")
			username := r.FormValue("username")
			chatRoom, err := createChatRoom(r.Context(), username, chatRoomName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(chatRoom)
		}
	})

	http.HandleFunc("/dms", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
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

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}
	log.Fatal(httpServer.ListenAndServe())
}

type httpAcceptType int

const (
	HTML httpAcceptType = iota
	JSON
	UNKNOWN
)

func getAcceptHeader(r *http.Request) httpAcceptType {
	rawAcceptHeader := r.Header.Get("Accept")
	acceptMimeTypes := strings.Split(rawAcceptHeader, ",")
	accpetsJSON := false
	accpetsHTML := false
	for _, acceptMimeType := range acceptMimeTypes {
		mimeTypeParts := strings.Split(acceptMimeType, ";")
		mimeType := mimeTypeParts[0]
		switch mimeType {
		case "text/html":
			accpetsHTML = true

		case "application/json":
			accpetsJSON = true
		}
	}

	if accpetsJSON && accpetsHTML {
		return UNKNOWN
	}
	if accpetsJSON {
		return JSON
	}
	if accpetsHTML {
		return HTML
	}

	return UNKNOWN
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
