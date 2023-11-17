package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	serverConn *websocket.Conn
	cancel     context.CancelFunc
	selfAddr   string
	context    context.Context
	textarea   textarea.Model
	messages   []string
	viewport   viewport.Model
	err        error
}

func (model) Init() tea.Cmd {
	return textarea.Blink
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
			sendMessage(text, m)
		}
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

func initialModel() model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	conn, r, err := websocket.Dial(ctx, "ws://localhost:8080", nil)
	addr := r.Request.RemoteAddr
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket.Dial: %v\n", err)
		os.Exit(1)
	}
	ta := textarea.New()
	ta.Placeholder = "Enter your message here and press enter to send it"
	ta.Focus()
	ta.Prompt = "| "
	ta.CharLimit = 180
	ta.SetWidth(70)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	vp := viewport.New(180, 8)
	return model{
		serverConn: conn,
		selfAddr:   addr,
		cancel:     cancel,
		context:    ctx,
		textarea:   ta,
		viewport:   vp,
		messages:   []string{},
		err:        nil,
	}
}

func sendMessage(text string, model model) {
	msg := message.Message{
		Type: message.Text,
		Data: text,
		From: model.selfAddr,
		To:   "[::1]:37770",
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error5: ", err)
	}
	model.serverConn.Write(model.context, websocket.MessageText, bytes)

}

func main() {
	m := initialModel()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
