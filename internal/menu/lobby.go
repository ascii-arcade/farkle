package menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/server"
	"github.com/ascii-arcade/farkle/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width     int
	height    int
	hostsName string

	creatingLobby bool

	errors    string
	menuModel menuModel
}

func newLobbyModel(menuModel menuModel, hostsName string) (lobbyModel, tea.Cmd) {
	wsClient = newWsClient(menuModel.logger)

	m := lobbyModel{
		width:     menuModel.width,
		height:    menuModel.height,
		menuModel: menuModel,
		hostsName: hostsName,
	}
	return m, m.Init()
}

func (m lobbyModel) Init() tea.Cmd {
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
			return m.menuModel, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tick(t)
			})
		case "enter":
			if currentLobby.Ready() {
				tui.RunFromLobby(currentLobby)
				return m.menuModel, nil
			}

			m.errors = "Please wait for at least two players to join before starting the game"

			return m, nil
		}
	case tick:
		if !m.creatingLobby && wsClient.IsConnected() {
			if err := wsClient.SendMessage(server.Message{
				Channel: server.ChannelLobby,
				Type:    server.MessageTypeCreate,
				Data:    m.hostsName,
				SentAt:  time.Now(),
			}); err != nil {
				m.errors = "Failed to create lobby"
				m.menuModel.logger.Error("Failed to send lobby message", "error", err)
				m.creatingLobby = false
				goto RETURN
			}

			m.creatingLobby = true
		}
	RETURN:
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
	lobbyStyle := lipgloss.NewStyle().Padding(1, 2).Margin(1).BorderStyle(lipgloss.NormalBorder()).Width(28)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width / 2)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.width / 2)

	if debug {
		fullPaneStyle = fullPaneStyle.BorderStyle(lipgloss.ASCIIBorder()).BorderForeground(lipgloss.Color("#0000ff")).Width(m.width - 2).Height(m.height - 3)
		controlsStyle = controlsStyle.Background(lipgloss.Color("#000066")).Foreground(lipgloss.Color("#ffffff"))
		errorsStyle = errorsStyle.Background(lipgloss.Color("#660000")).Foreground(lipgloss.Color("#ffffff"))
	}

	if currentLobby == nil {
		msg := "Connecting to server..."
		if wsClient.IsConnected() {
			msg = "Loading lobby..."
		}
		if m.errors != "" {
			msg = m.errors + "\nPress enter to go back to menu"
		}

		return fullPaneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render(msg),
			),
		)
	}

	lobbyContent := []string{}

	for i, player := range currentLobby.Players {
		if player != nil && player.Host {
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s (Host)", i+1, player.Name))
			continue
		}

		if player == nil {
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) ...", i+1))
			continue
		}

		lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s", i+1, player.Name))
	}

	lobbyNamePane := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render("Lobby: " + currentLobby.Name + "\n")

	lobbyPane := lobbyStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lobbyNamePane,
			strings.Join(lobbyContent, "\n"),
			lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render(""),
		),
	)

	controlsPane := lipgloss.JoinHorizontal(
		lipgloss.Left,
		controlsStyle.Render("esc to go back to menu, enter to start game"),
		errorsStyle.Render(m.errors),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		fullPaneStyle.Render(
			lobbyPane,
		),
		controlsPane,
	)
}
