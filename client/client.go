package main

import (
	"adisuper94/turboguac/turbosdk"
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func WsListen(t turboTUIClient) tea.Cmd {
	return func() tea.Msg {
		t.tgc.WSListen(t.wsMessageChan)
		return tea.Quit
	}
}

var (
	columnStyle = lipgloss.NewStyle().
		// Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241"))
	focusedStyle = lipgloss.NewStyle().
		// Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("41"))
)

func (t turboTUIClient) Init() tea.Cmd {
	return tea.Batch(t.chat.Init(), t.onlineUsers.Init(), t.myChatRooms.Init(), WsListen(t), ReadChannel(t))
}

func ReadChannel(t turboTUIClient) tea.Cmd {
	return func() tea.Msg {
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
}

func (t turboTUIClient) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var m tea.Model
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = t.resizeChat(msg.Width, msg.Height-2)
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
	case OpenChatMsg:
		t.focucedUI = Chat
		m, cmd = t.chat.Update(msg)
		t.chat = m.(chatModel)
	case OnlineUsersMsg:
		m, cmd = t.onlineUsers.Update(msg)
		t.onlineUsers = m.(onlineUserModel)
	case MyChatRoomsMsg:
		m, cmd = t.myChatRooms.Update(msg)
		t.myChatRooms = m.(myChatRoomsModel)
	case IncomingChatMsg:
		m, cmd = t.chat.Update(msg)
		cmd = tea.Batch(cmd, ReadChannel(t))
		t.chat = m.(chatModel)
	}
	return t, cmd
}

func (t turboTUIClient) View() string {
	chatBoxView := t.chat.View()
	onlineUsersView := t.onlineUsers.View()
	myChatRoomsView := t.myChatRooms.View()
	switch t.focucedUI {
	case OnlineUsers:
		onlineUsersView = focusedStyle.Render(onlineUsersView)
		myChatRoomsView = columnStyle.Render(myChatRoomsView)
		chatBoxView = columnStyle.Render(chatBoxView)
	case MyChatRooms:
		myChatRoomsView = focusedStyle.Render(myChatRoomsView)
		onlineUsersView = columnStyle.Render(onlineUsersView)
		chatBoxView = columnStyle.Render(chatBoxView)
	case Chat:
		chatBoxView = focusedStyle.Render(chatBoxView)
		onlineUsersView = columnStyle.Render(onlineUsersView)
		myChatRoomsView = columnStyle.Render(myChatRoomsView)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.JoinVertical(lipgloss.Top, onlineUsersView, myChatRoomsView), chatBoxView)
}

type ChatWindowsResizeMsg struct {
	Width  int
	Height int
}

type OnlineUserWindowsResizeMsg struct {
	Width  int
	Height int
}

type MyChatRoomsWindowsResizeMsg struct {
	Width  int
	Height int
}

func (t turboTUIClient) resizeChat(width, height int) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ChatWindowsResizeMsg{Width: 2 * width / 3, Height: height}
		},
		func() tea.Msg {
			return OnlineUserWindowsResizeMsg{Width: width / 3, Height: height / 2}
		},
		func() tea.Msg {
			return MyChatRoomsWindowsResizeMsg{Width: width / 3, Height: height / 2}
		},
	)
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
	var username string
	fmt.Print("Enter your username: ")
	fmt.Scanln(&username)
	t.tgc, err = turbosdk.NewTurboGuacClient(context.Background(), username, "localhost:8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewTurboGuacClient() failed in turboTUIClient: \n %v", err)
		os.Exit(1)
	}
	t.focucedUI = Chat
	totalWidth := 100
	totalHeight := 40
	t.chat = initialChatModel(t.tgc, 2*totalWidth/3, totalHeight)
	t.onlineUsers = InitalOnlineUserModel(t.tgc, totalWidth/3, totalHeight/2)
	t.myChatRooms = InitialMyChatRoomsModel(t.tgc, totalWidth/3, totalHeight/2)
	t.wsMessageChan = make(chan turbosdk.IncomingChat)
	t.cachedChatRooms = CachedChatRooms{ChatRoomMap: map[uuid.UUID]CachedChatRoom{}}
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
