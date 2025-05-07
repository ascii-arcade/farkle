package menu

import (
	"slices"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	fullPaneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center)
	lobbyStyle := lipgloss.NewStyle().Margin(1, 2).Padding(1, 2).Border(lipgloss.NormalBorder(), true)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).AlignHorizontal(lipgloss.Center)

	l := []string{}
	l = append(l, "Lobby: "+m.name)
	l = append(l, "Players: "+strconv.Itoa(len(m.players)))

	for i, player := range m.players {
		if i == 0 {
			player = player + " (Host)"
		}
		l = append(l, strconv.Itoa(i+1)+") "+player)
	}

	errors := ""
	if slices.Contains(m.players, "") {
		errors = "Waiting for players to join..."
	}

	lobbyRender := lobbyStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			l...,
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		fullPaneStyle.Render(
			lobbyRender,
		),
		errorsStyle.Render(errors),
		controlsStyle.Render("ESC to exit, Enter to start the game"),
	)
}
