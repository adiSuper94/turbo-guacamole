package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type myChatRoomsModel struct {
	tgc         *turbosdk.TurboGuacClient
	myChatRooms list.Model
}

type MyChatRoomsMsg []turbosdk.ChatRoom

type ChatRoomItem struct {
	Name string
	ID   uuid.UUID
}

func (o ChatRoomItem) Title() string {
	return o.Name
}

func (o ChatRoomItem) FilterValue() string {
	return o.Name
}

func (o ChatRoomItem) Description() string {
	return ""
}

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
	return UpdateMyChatRooms(m)
}

func (m myChatRoomsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.myChatRooms, cmd = m.myChatRooms.Update(msg)
	switch msg := msg.(type) {
	case MyChatRoomsWindowsResizeMsg:
		m.myChatRooms.SetWidth(msg.Width)
		m.myChatRooms.SetHeight(msg.Height)
	case MyChatRoomsMsg:
		newChatRooms := []list.Item{}
		for _, chatRoom := range msg {
			newChatRooms = append(newChatRooms, ChatRoomItem{Name: chatRoom.Name, ID: chatRoom.ID})
		}
		m.myChatRooms.SetItems(newChatRooms)
	case tea.KeyMsg:
		switch msg.String() {
		case "R":
			cmd = UpdateMyChatRooms(m)
		case "enter":
			selectedChatRoom := m.myChatRooms.SelectedItem().(ChatRoomItem)
			return m, OpenChat(turbosdk.ChatRoom{ID: selectedChatRoom.ID, Name: selectedChatRoom.Name})
		}
	}
	return m, cmd
}

func (m myChatRoomsModel) View() string {
	return m.myChatRooms.View()
}

func InitialMyChatRoomsModel(tgc *turbosdk.TurboGuacClient, width int, height int) myChatRoomsModel {
	m := myChatRoomsModel{tgc: tgc}
	m.myChatRooms = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.myChatRooms.Title = "My Chat Rooms"
	return m
}
