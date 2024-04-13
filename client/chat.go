package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type chatModel struct {
	tgc        *turbosdk.TurboGuacClient
	textbox    textarea.Model
	messages   viewport.Model
	activeChat turbosdk.ChatRoom
}

type OpenChatMsg turbosdk.ChatRoom

type IncomingChatMsg turbosdk.IncomingChat

func OpenChat(chatRoom turbosdk.ChatRoom) tea.Cmd {
	return func() tea.Msg {
		return OpenChatMsg(chatRoom)
	}
}

func (m chatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var taCmd, vpCmd tea.Cmd
	m.textbox, taCmd = m.textbox.Update(msg)
	m.messages, vpCmd = m.messages.Update(msg)
	switch msg := msg.(type) {
	case ChatWindowsResizeMsg:
		m.textbox.SetWidth(msg.Width)
		m.textbox.SetHeight(msg.Height / 3)
		m.messages.Width = msg.Width
		m.messages.Height = 2 * msg.Height / 3
	case IncomingChatMsg:
		if m.activeChat.ID == msg.To {
			m.messages.SetContent(fmt.Sprintf("%s\n%s: %s", m.messages.View(), msg.From, msg.Message))
			m.messages.GotoBottom()
		}
	case OpenChatMsg:
		m.activeChat = turbosdk.ChatRoom(msg)
		m.textbox.Reset()
		m.messages.SetContent("")
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			chatMessage := m.textbox.Value()
			m.textbox.Reset()
			m.messages.SetContent(fmt.Sprintf("%s\nYou: %s", m.messages.View(), chatMessage))
			m.messages.GotoBottom()
			err := m.tgc.SendMessage(chatMessage, m.activeChat.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SendMessage() failed in chatModel: \n %v", err)
				return m, tea.Quit
			}
		}
	}
	return m, tea.Batch(taCmd, vpCmd)
}

func (m chatModel) View() string {
	return fmt.Sprintf("%s\n\n%s", m.messages.View(), m.textbox.View())
}

func initialChatModel(tgc *turbosdk.TurboGuacClient, width int, height int) chatModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message here ..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.CharLimit = 280
	ta.SetWidth(width)
	ta.SetHeight(height / 3)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	messagesCanvas := viewport.New(width, 2*height/3)
	return chatModel{tgc: tgc, textbox: ta, messages: messagesCanvas}
}
