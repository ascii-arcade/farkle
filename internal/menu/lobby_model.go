package menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width  int
	height int

	errors    string
	menuModel menuModel
}

func newLobbyModel(menuModel menuModel, name string, hostName string) lobbyModel {
	lm := lobbyModel{
		width:     menuModel.width,
		height:    menuModel.height,
		menuModel: menuModel,
	}

	currentLobby = lobby.NewLobby(name, hostName)
	// b, err := currentLobby.ToBytes()
	// if err != nil {
	// 	return lm
	// }
	lm.errors = "TEST"

	return lm
}

func (m lobbyModel) Init() tea.Cmd {
	return nil
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m.menuModel, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tick(t)
			})
		case "enter":
			if currentLobby.Ready() {
				tui.RunFromLobby(currentLobby)
				return m.menuModel, nil
			}

			return m, nil
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
	fullPaneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height-1).Align(lipgloss.Center, lipgloss.Center)
	lobbyStyle := lipgloss.NewStyle().Padding(1, 2).Margin(1).BorderStyle(lipgloss.NormalBorder())
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Margin(1)

	if debug {
		fullPaneStyle = fullPaneStyle.BorderStyle(lipgloss.ASCIIBorder()).BorderForeground(lipgloss.Color("#0000ff")).Width(m.width - 2).Height(m.height - 3)
		controlsStyle = controlsStyle.Background(lipgloss.Color("#0000ff")).Foreground(lipgloss.Color("#ffffff"))
		errorsStyle = errorsStyle.BorderStyle(lipgloss.ASCIIBorder()).BorderForeground(lipgloss.Color("#0000ff")).Margin(0)
	}

	lobbyContent := []string{}

	for i, player := range currentLobby.Players {
		if player != nil && player.Host {
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s (Host)", i, player.Name))
			continue
		}

		if player == nil {
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) (Waiting for player to join...)", i))
			continue
		}

		lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s", i, player.Name))
	}

	lobbyPane := lobbyStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render("Lobby: "+currentLobby.Name),
			strings.Join(lobbyContent, "\n"),
		),
	)

	controlsPane := lipgloss.JoinHorizontal(
		lipgloss.Left,
		controlsStyle.Render("ESC to exit, Enter to start the game"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		fullPaneStyle.Render(
			lobbyPane,
			errorsStyle.Render(m.errors),
		),
		controlsPane,
	)
}
