package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type myChatRoomsModel struct {
	tgc         *turbosdk.TurboGuacClient
	myChatRooms []turbosdk.ChatRoom
	highlighted int
}

type MyChatRoomsMsg []turbosdk.ChatRoom

func UpdateMyChatRooms(m myChatRoomsModel) tea.Cmd {
	return func() tea.Msg {
		myChatRooms, err := m.tgc.GetMyChatRooms()
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetMyChatRooms() REFRESH failed in myChatRoomsModel: \n %v", err)
			return tea.Quit
		}
		return MyChatRoomsMsg(myChatRooms)
	}
}

func (m myChatRoomsModel) Init() tea.Cmd {
	m.highlighted = 0
	return UpdateMyChatRooms(m)
}

func (m myChatRoomsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case MyChatRoomsMsg:
		m.myChatRooms = msg
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
			cmd = UpdateMyChatRooms(m)
		case "enter":
			selectedChatRoom := m.myChatRooms[m.highlighted]
			return m, OpenChat(turbosdk.ChatRoom{ID: selectedChatRoom.ID, Name: selectedChatRoom.Name})
		}
	}
	return m, cmd
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
