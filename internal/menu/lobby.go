package menu

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type lobbyModel struct {
	width  int
	height int

	name      string
	players   []string
	menuModel menuModel
}

type lobbyTick time.Time

func newLobbyModel(menuModel menuModel, name string, host string, playerCount int) lobbyModel {
	players := make([]string, playerCount)
	players[0] = host

	return lobbyModel{
		width:  menuModel.width,
		height: menuModel.height,

		name:      name,
		players:   players,
		menuModel: menuModel,
	}
}

func (m lobbyModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return lobbyTick(t)
	})
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m.menuModel, nil
		}
	case lobbyTick:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return lobbyTick(t)
		})
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, nil
}

func (m lobbyModel) View() string {
	return "Lobby Name: " + m.name + "\n" +
		"Players: " + strconv.Itoa(len(m.players)) + "\n" +
		"Press Ctrl+C or q to quit."
}
