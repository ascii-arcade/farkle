package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/server"
	"github.com/ascii-arcade/farkle/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width  int
	height int

	menuModel menuModel

	lobby *lobby.Lobby
}

func newLobbyModel(menuModel menuModel, name string, hostName string, playerCount int) lobbyModel {
	lm := lobbyModel{
		width:     menuModel.width,
		height:    menuModel.height,
		menuModel: menuModel,
	}

	l := lobby.NewLobby(name, hostName)
	b, err := l.ToBytes()
	if err != nil {
		return lm
	}

	lm.lobby = l

	wsClient.SendMessage(server.Message{
		Channel: server.ChannelLobby,
		Type:    server.MessageTypeCreate,
		Data:    b,
	})

	return lm
}

func (m lobbyModel) Init() tea.Cmd {
	lobbyData, err := m.lobby.ToBytes()
	if err != nil {
		// TODO: need to handle this error
		return nil
	}
	if err := wsClient.SendMessage(server.Message{
		Channel: server.ChannelLobby,
		Type:    server.MessageTypeCreate,
		Data:    lobbyData,
	}); err != nil {
		// TODO: need to handle this error
		return nil
	}

	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tick(t)
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
		case "enter":

			return m.menuModel, nil
		}
	case tick:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tick(t)
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

	l := []string{}
	l = append(l, "Lobby: "+m.lobby.Name)

	for name, player := range m.lobby.Players {
		if player.Host {
			l = append(l, name+" (Host)")
			continue
		}

		if player == nil || player.Name == "" {
			l = append(l, name+" (Waiting for player to join...)")
			continue
		}
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
		controlsStyle.Render("ESC to exit, Enter to start the game"),
	)
}

func missingPlayers(players map[string]*tui.Player) int {
	missing := len(players)
	for _, player := range players {
		if player == nil || player.Name == "" {
			missing--
		}
	}
	return missing
}
