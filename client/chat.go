package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	case IncomingChatMsg:
		log.Println("IncomingChatMsg")
		if m.activeChat.ID == msg.To {
			m.messages.SetContent(fmt.Sprintf("%s\n%s: %s", m.messages.View(), msg.From, msg.Message))
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

func initialChatModel(tgc *turbosdk.TurboGuacClient) chatModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message here ..."
	ta.Focus()
	ta.Prompt = "> "
	ta.CharLimit = 100
	ta.SetWidth(70)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	messagesCanvas := viewport.New(180, 8)
	return chatModel{tgc: tgc, textbox: ta, messages: messagesCanvas}
}
