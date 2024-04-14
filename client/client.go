package main

import (
	"adisuper94/turboguac/turbosdk"
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type UI int

const (
	OnlineUsers UI = iota
	MyChatRooms
	Chat
)

type turboTUIClient struct {
	tgc         *goclient.TurboGuacClient
	focucedUI   UI
	chat        chatModel
	onlineUsers onlineUserModel
	myChatRooms myChatRoomsModel
	ctx         context.Context
}

func (t turboTUIClient) Init() tea.Cmd {
	return tea.Batch(t.chat.Init(), t.onlineUsers.Init(), t.myChatRooms.Init())
}

func (t turboTUIClient) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var m tea.Model
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return t, tea.Quit
		case tea.KeyTab:
			t.focucedUI = t.getNextFocus()
		default:
			switch t.focucedUI {
			case OnlineUsers:
				m, cmd = t.onlineUsers.Update(msg)
				t.onlineUsers = m.(onlineUserModel)
			case MyChatRooms:
				m, cmd = t.myChatRooms.Update(msg)
				t.myChatRooms = m.(myChatRoomsModel)
			case Chat:
				m, cmd = t.chat.Update(msg)
				t.chat = m.(chatModel)
			}
		}
	case OnlineUsersMsg:
		m, cmd = t.onlineUsers.Update(msg)
		t.onlineUsers = m.(onlineUserModel)
	case MyChatRoomsMsg:
		m, cmd = t.myChatRooms.Update(msg)
		t.myChatRooms = m.(myChatRoomsModel)
	}
	return t, cmd
}

func (t turboTUIClient) View() string {
	return fmt.Sprintf("%s\n%s\n%s", t.onlineUsers.View(), t.myChatRooms.View(), t.chat.View())
}

func (t turboTUIClient) getNextFocus() UI {
	switch t.focucedUI {
	case OnlineUsers:
		return MyChatRooms
	case MyChatRooms:
		return Chat
	case Chat:
		return OnlineUsers
	}
	return Chat
}

func initialMainModel() turboTUIClient {
	var err error
	t := turboTUIClient{}
	t.tgc, err = goclient.NewTurboGuacClient(context.Background(), "aditya", "localhost:8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewTurboGuacClient() failed in turboTUIClient: \n %v", err)
		os.Exit(1)
	}
	t.focucedUI = Chat
	t.chat = initialChatModel(t.tgc)
	t.onlineUsers = onlineUserModel{tgc: t.tgc}
	t.myChatRooms = myChatRoomsModel{tgc: t.tgc}
	return t
}

func main() {
	ttc := initialMainModel()
	p := tea.NewProgram(ttc)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in main(): \n %v", err)
		os.Exit(1)
	}
}
