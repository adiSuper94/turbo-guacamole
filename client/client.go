package main

import (
	"adisuper94/turboguac/turbosdk"
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type UI int

const (
	OnlineUsers UI = iota
	MyChatRooms
	Chat
)

type ChatMessage struct {
	From    string
	Message string
}

type CachedChatRoom struct {
	ChatRoomId   uuid.UUID
	ChatRoomName string
	Messages     []ChatMessage
}

type CachedChatRooms struct {
	ChatRoomMap map[uuid.UUID]CachedChatRoom
}

type turboTUIClient struct {
	tgc             *turbosdk.TurboGuacClient
	focucedUI       UI
	chat            chatModel
	onlineUsers     onlineUserModel
	myChatRooms     myChatRoomsModel
	cachedChatRooms CachedChatRooms
	wsMessageChan   chan turbosdk.IncomingChat
}

func wsListen(t turboTUIClient) tea.Cmd {
	return func() tea.Msg {
		t.tgc.WSListen(t.wsMessageChan)
		return tea.Quit
	}
}

func (t turboTUIClient) Init() tea.Cmd {
	return tea.Batch(t.chat.Init(), t.onlineUsers.Init(), t.myChatRooms.Init(), wsListen(t), t.readChannel)
}

func (t turboTUIClient) readChannel() tea.Msg {
	incomingChat := <-t.wsMessageChan
	chatRoomId := incomingChat.To
	cachedChatRoom, ok := t.cachedChatRooms.ChatRoomMap[chatRoomId]
	if !ok {
		// TODO: Fetch chat room name details from server
		cachedChatRoom = CachedChatRoom{
			ChatRoomId:   chatRoomId,
			ChatRoomName: "Unknown",
			Messages:     []ChatMessage{},
		}
	}
	cachedChatRoom.Messages = append(cachedChatRoom.Messages, ChatMessage{
		From:    incomingChat.From,
		Message: incomingChat.Message})
	t.cachedChatRooms.ChatRoomMap[chatRoomId] = cachedChatRoom
	return IncomingChatMsg(incomingChat)
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
	case IncomingChatMsg:
		m, cmd = t.chat.Update(msg)
		cmd = tea.Batch(cmd, t.readChannel)
		t.chat = m.(chatModel)
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
	t.tgc, err = turbosdk.NewTurboGuacClient(context.Background(), "aditya", "localhost:8080")
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
