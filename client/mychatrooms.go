package main

import (
	"adisuper94/turboguac/goclient"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type myChatRoomsModel struct {
	tgc         *goclient.TurboGuacClient
	myChatRooms []goclient.ChatRoom
	highlighted int
}

func (m myChatRoomsModel) Init() tea.Cmd {
	var err error
	m.myChatRooms, err = m.tgc.GetMyChatRooms()
	m.highlighted = 0
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetMyChatRooms() failed in myChatRoomsModel: \n %v", err)
		return tea.Quit
	}
	return nil
}

func (m myChatRoomsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.highlighted < len(m.myChatRooms)-1 {
				m.highlighted++
			}
		case "k", "up":
			if m.highlighted > 0 {
				m.highlighted--
			}
		case "R":
			m.myChatRooms, err = m.tgc.GetMyChatRooms()
			if err != nil {
				fmt.Fprintf(os.Stderr, "GetMyChatRooms() REFRESH failed in myChatRoomsModel: \n %v", err)
				return m, tea.Quit
			}
		case "enter":
			// join the highlighted chat room
		}
	}
	return m, nil
}

func (m myChatRoomsModel) View() string {
	s := "My Chat Rooms\n\n"
	for i, chatRoom := range m.myChatRooms {
		if i == m.highlighted {
			s += fmt.Sprintf("%s\n", chatRoom.Name)
		} else {
			s += fmt.Sprintf("  %s\n", chatRoom.Name)
		}
	}
	return s
}
