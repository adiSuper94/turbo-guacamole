package main

import (
	"adisuper94/turboguac/turbosdk"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type onlineUserModel struct {
	tgc         *turbosdk.TurboGuacClient
	onlineUsers list.Model
}

type OnlineUsersMsg []string

type OnlineUserItem struct {
	username string
}

func (o OnlineUserItem) Title() string {
	return o.username
}

func (o OnlineUserItem) FilterValue() string {
	return o.username
}

func (o OnlineUserItem) Description() string {
	return ""
}

func UpdateOnlineUsers(m onlineUserModel) tea.Cmd {
	return func() tea.Msg {
		onlineUsers, err := m.tgc.GetOnlineUsers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetOnlineUsers() REFRESH failed in onlineUserModel: \n %v", err)
			return tea.Quit
		}
		onlinUsersExcludingMe := []string{}
		me := m.tgc.GetUsername()
		for _, user := range onlineUsers {
			if user != me {
				onlinUsersExcludingMe = append(onlinUsersExcludingMe, user)
			}
		}
		return OnlineUsersMsg(onlinUsersExcludingMe)
	}
}

func (m onlineUserModel) Init() tea.Cmd {
	return UpdateOnlineUsers(m)
}

func (m onlineUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.onlineUsers, cmd = m.onlineUsers.Update(msg)
	switch msg := msg.(type) {
	case ChatWindowsResizeMsg:
		m.onlineUsers.SetWidth(msg.Width)
		m.onlineUsers.SetHeight(msg.Height)
	case OnlineUsersMsg:
		onlineUsers := []list.Item{}
		for _, username := range msg {
			onlineUsers = append(onlineUsers, OnlineUserItem{username: username})
		}
		cmd = tea.Batch(cmd, m.onlineUsers.SetItems(onlineUsers))
	case tea.KeyMsg:
		switch msg.String() {
		case "R":
			cmd = tea.Batch(cmd, UpdateOnlineUsers(m))
		case "enter":
			selectedUser := m.onlineUsers.SelectedItem().(OnlineUserItem).username
			chatRoomId, err := m.tgc.StartDM(selectedUser)
			if err != nil {
				fmt.Fprintf(os.Stderr, "StartDM() failed in onlineUserModel: \n %v", err)
				return m, tea.Quit
			}
			return m, tea.Batch(cmd, OpenChat(turbosdk.ChatRoom{ID: chatRoomId, Name: selectedUser}))
		}
	}
	return m, cmd
}

func (m onlineUserModel) View() string {
	return m.onlineUsers.View()
}

func InitalOnlineUserModel(tgc *turbosdk.TurboGuacClient, width int, height int) onlineUserModel {
	m := onlineUserModel{tgc: tgc}
	m.onlineUsers = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.onlineUsers.Title = "Online Users"
	return m
}
