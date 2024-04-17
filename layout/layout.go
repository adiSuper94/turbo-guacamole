package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Layout struct {
	chatWindowWidth   int
	chatWindowHeight  int
	onlineUsersWidth  int
	onlineUsersHeight int
	mychatroomsWidth  int
	mychatroomsHeight int
	twidth            int
	theight           int
	chatWindow        viewport.Model
	onlineUsers       viewport.Model
	mychatrooms       viewport.Model
}

var style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("241"))

func initalizeLayout() Layout {
	layout := Layout{
		chatWindowWidth:   0,
		chatWindowHeight:  0,
		onlineUsersWidth:  0,
		onlineUsersHeight: 0,
		mychatroomsWidth:  0,
		mychatroomsHeight: 0,
		chatWindow:        viewport.New(1, 1),
		onlineUsers:       viewport.New(1, 1),
		mychatrooms:       viewport.New(1, 1),
	}
	return layout
}

func (layout Layout) Init() tea.Cmd {
	return nil
}

func (l Layout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd1, cmd2, cmd3 tea.Cmd
	l.chatWindow, cmd1 = l.chatWindow.Update(msg)
	l.onlineUsers, cmd2 = l.onlineUsers.Update(msg)
	l.mychatrooms, cmd3 = l.mychatrooms.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		twidth := msg.Width
		theight := msg.Height
		l.resize(twidth, theight)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return l, tea.Quit
		}
	}
	return l, tea.Batch(cmd1, cmd2, cmd3)
}

func (layout Layout) View() string {
	layout.chatWindow.Style = style
	layout.onlineUsers.Style = style
	layout.mychatrooms.Style = style
	chatWindow := layout.chatWindow.View()
	onlineUsers := layout.onlineUsers.View()
	mychatrooms := layout.mychatrooms.View()
	return lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.JoinVertical(lipgloss.Top, mychatrooms, onlineUsers), chatWindow)
}

func (l Layout) resize(width int, height int) Layout {
	l.chatWindowWidth = (2 * width / 3) - 1
	l.chatWindowHeight = height
	l.mychatroomsWidth = width / 3
	l.mychatroomsHeight = height / 2
	l.onlineUsersWidth = width / 3
	l.onlineUsersHeight = height / 2
	l.chatWindow = viewport.New(l.chatWindowWidth, l.chatWindowHeight)
	l.mychatrooms = viewport.New(l.mychatroomsWidth, l.mychatroomsHeight)
	l.onlineUsers = viewport.New(l.onlineUsersWidth, l.onlineUsersHeight)
	return l
}

func main() {
	l := initalizeLayout()
	p := tea.NewProgram(l)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in main(): \n %v", err)
		os.Exit(1)
	}
}
