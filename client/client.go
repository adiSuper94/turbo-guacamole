package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"turboGuac/message"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"nhooyr.io/websocket"
)

type model struct {
	serverConn       *websocket.Conn
	cancel           context.CancelFunc
	selfAddr         string
	toAddr           string
	context          context.Context
	textarea         textarea.Model
	incomingMessages chan string
	messages         []string
	viewport         viewport.Model
	err              error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, receiveMessages(m), listenForIncomingMessages(m))
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, tiCmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			text := m.textarea.Value()
			m.messages = append(m.messages, "You: "+text)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			sendTextMessage(text, m)
		}
	case IncomingMessage:
		incomingMsg := string(msg)
		m.messages = append(m.messages, m.toAddr+": "+incomingMsg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		listenForIncomingMessages(m)
		return m, tea.Batch(tiCmd, vpCmd, listenForIncomingMessages(m))
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

// View implements tea.Model.
func (m model) View() string {

	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func initialModel(conn *websocket.Conn, ctx context.Context, addr string, toAddr string, cancel context.CancelFunc) model {
	ta := textarea.New()
	ta.Placeholder = "Enter your message here and press enter to send it"
	ta.Focus()
	ta.Prompt = "| "
	ta.CharLimit = 180
	ta.SetWidth(70)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	vp := viewport.New(180, 8)
	return model{
		serverConn:       conn,
		selfAddr:         addr,
		toAddr:           toAddr,
		cancel:           cancel,
		context:          ctx,
		textarea:         ta,
		viewport:         vp,
		incomingMessages: make(chan string),
		messages:         []string{},
		err:              nil,
	}
}

func sendTextMessage(text string, model model) {
	msg := message.Message{
		Type: message.Text,
		Data: text,
		From: model.selfAddr,
		To:   model.toAddr,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error5: ", err)
	}
	model.serverConn.Write(model.context, websocket.MessageText, bytes)

}

func login(conn *websocket.Conn, ctx context.Context) string {
	loginMsg := message.Message{
		Type: message.Login,
		Data: "",
		From: "me",
		To:   "God",
	}
	bytes, err := json.Marshal(loginMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: login message is malformed", err)
	}
	conn.Write(ctx, websocket.MessageText, bytes)
	_, respBytes, err := conn.Read(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while reading login resposne. \n err: ", err)
	}
	var respMsg message.Message
	err = json.Unmarshal(respBytes, &respMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while unmarshalling login resposne. \n err: ", err)
	}
	return respMsg.Data
}

type IncomingMessage string

func listenForIncomingMessages(model model) tea.Cmd {
	return func() tea.Msg {
		msg := <-model.incomingMessages
		return IncomingMessage(msg)
	}
}

func receiveMessages(m model) tea.Cmd {
	return func() tea.Msg {
		for {
			_, rawMessageBytes, err := m.serverConn.Read(m.context)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error, while reading incoming messags", err)
				continue
			}
			var msg message.Message
			err = json.Unmarshal(rawMessageBytes, &msg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error, unmarshalling incoming message", err)
				continue
			}
			if msg.Type == message.Text {
				m.incomingMessages <- msg.Data
			}
		}
	}
}

func selectFriendIp(addr string) string {
	var friendIp string
	for {
		resp, err := http.Get("http://localhost:8080/online-users")
		if err != nil {
			fmt.Println(err)
			continue
		}
		onlineUsers := make([]string, 10)
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(respBytes, &onlineUsers)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(onlineUsers)
		var idx int
		_, err = fmt.Scanf("%d", &idx)
		if err != nil || idx < 0 || idx >= len(onlineUsers) {
			fmt.Println(err, "You fool! Enter a valid index.")
			continue
		}
		friendIp = onlineUsers[idx]
		if friendIp == addr {
			fmt.Println("You fool! You can't talk to yourself.")
			continue
		}
		break
	}
	return friendIp
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket.Dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(websocket.StatusNormalClosure, "the sky is falling")
	addr := login(conn, ctx)
	friendIp := selectFriendIp(addr)

	m := initialModel(conn, ctx, addr, friendIp, cancel)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
