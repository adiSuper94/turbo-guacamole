package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type onlineUserModel struct {
	tgc         *turbosdk.TurboGuacClient
	onlineUsers []string
	highlighted int
}

type OnlineUsersMsg []string

func UpdateOnlineUsers(m onlineUserModel) tea.Cmd {
	return func() tea.Msg {
		onlineUsers, err := m.tgc.GetOnlineUsers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetOnlineUsers() REFRESH failed in onlineUserModel: \n %v", err)
			return tea.Quit
		}
		return OnlineUsersMsg(onlineUsers)
	}
}

func (m onlineUserModel) Init() tea.Cmd {
	m.highlighted = 0
	return UpdateOnlineUsers(m)
}

func (m onlineUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case OnlineUsersMsg:
		m.onlineUsers = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.highlighted < len(m.onlineUsers)-1 {
				m.highlighted++
			}
		case "k", "up":
			if m.highlighted > 0 {
				m.highlighted--
			}
		case "R":
			cmd = UpdateOnlineUsers(m)
		case "enter":
			selectedUser := m.onlineUsers[m.highlighted]
			chatRoomId, err := m.tgc.StartDM(selectedUser)
			if err != nil {
				fmt.Fprintf(os.Stderr, "StartDM() failed in onlineUserModel: \n %v", err)
				return m, tea.Quit
			}
			return m, OpenChat(turbosdk.ChatRoom{ID: chatRoomId, Name: selectedUser})
		}
	}
	return m, cmd
}

func (m onlineUserModel) View() string {
	s := "Online Users\n\n"
	for i, user := range m.onlineUsers {
		if i == m.highlighted {
			s += fmt.Sprintf("> %s\n", user)
		} else {
			s += fmt.Sprintf("%s\n", user)
		}
	}
	return s
}
