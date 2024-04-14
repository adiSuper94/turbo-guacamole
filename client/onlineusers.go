package main

import (
	"adisuper94/turboguac/goclient"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type onlineUserModel struct {
	tgc         *goclient.TurboGuacClient
	onlineUsers []string
	highlighted int
}

func (m onlineUserModel) Init() tea.Cmd {
	var err error
	m.onlineUsers, err = m.tgc.GetOnlineUsers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetOnlineUsers() failed in onlineUserModel: \n %v", err)
		return tea.Quit
	}
	m.highlighted = 0
	return nil
}

func (m onlineUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	switch msg := msg.(type) {
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
			m.onlineUsers, err = m.tgc.GetOnlineUsers()
			if err != nil {
				fmt.Fprintf(os.Stderr, "GetOnlineUsers() REFRESH failed in onlineUserModel: \n %v", err)
				return m, tea.Quit
			}
		case "enter":
			// send a private message to the highlighted user
		}
	}
	return m, nil
}

func (m onlineUserModel) View() string {
	s := "Online Users\n\n"
	for i, user := range m.onlineUsers {
		if i == m.highlighted {
			s += fmt.Sprintf("%s\n", user)
		} else {
			s += fmt.Sprintf("  %s\n", user)
		}
	}
	return s
}
