package menu

import (
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type joinGameModel struct {
	width      int
	height     int
	focusIndex int

	menuModel menuModel
	logger    *slog.Logger
	debug     bool
}

func newJoinGameModel(menuModel menuModel) joinGameModel {
	return joinGameModel{
		width:      menuModel.width,
		height:     menuModel.height,
		focusIndex: 0,
		menuModel:  menuModel,
		logger:     menuModel.logger.With("component", "join_game"),
		debug:      menuModel.debug,
	}
}

func (m joinGameModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

func (m joinGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	}

	return m, nil
}

func (m joinGameModel) View() string {
	paneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center)

	if m.height < 15 || m.width < 100 {
		return paneStyle.Render("Window too small, please resize to something larger.")
	}

	lobbiesPaneStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		BorderStyle(lipgloss.NormalBorder()).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)

	lobbyNames := make([]string, 0, len(lobbies))
	for i, lobby := range lobbies {
		prefix := "  "
		if i == m.focusIndex {
			prefix = "> "
		}
		lobbyNames = append(lobbyNames, prefix+lobby.Name)
	}

	if len(lobbyNames) == 0 {
		lobbyNames = append(lobbyNames, "No lobbies available")
	}

	lobbiesPane := lipgloss.JoinVertical(lipgloss.Center, lobbiesPaneStyle.Render(lobbyNames...))

	return paneStyle.Render(lobbiesPane)
}
