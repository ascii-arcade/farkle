package menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width  int
	height int
	code   string

	creatingLobby bool
	joiningLobby  bool

	errors    string
	menuModel menuModel
}

func newLobbyModel(playerName string, code string, joining bool) (lobbyModel, tea.Cmd) {
	wsClient = newWsClient(playerName)

	m := lobbyModel{
		joiningLobby: joining,
		code:         code,
	}

	return m, m.Init()
}

func (m lobbyModel) Init() tea.Cmd {
	sizeCmd := tea.WindowSize()
	tickCmd := tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tick(t)
	})
	return tea.Batch(sizeCmd, tickCmd)
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			wsClient.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeLeave,
				SentAt:  time.Now(),
			})

			currentLobby = nil

			return newMenu(), m.menuModel.Init()
		case "enter":
			// if currentLobby.Ready() {
			// 	tui.RunFromLobby(currentLobby)
			// 	return m.menuModel, nil
			// }

			// m.errors = "Please wait for at least two players to join before starting the game"

			return m, nil
		}
	case tick:
		if currentLobby == nil && !m.joiningLobby && wsClient.IsConnected() {
			if err := wsClient.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeCreate,
				SentAt:  time.Now(),
			}); err != nil {
				m.errors = "Failed to create lobby"
				logger.Error("Failed to send lobby message", "error", err)
				m.creatingLobby = false
				goto RETURN
			}

			m.creatingLobby = true
		}

		if currentLobby == nil && m.joiningLobby && wsClient.IsConnected() {
			if err := wsClient.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeJoin,
				Data:    m.code,
				SentAt:  time.Now(),
			}); err != nil {
				m.errors = "Failed to join lobby"
				logger.Error("Failed to send lobby message", "error", err)
				goto RETURN
			}
		}

		if currentLobby != nil {
			m.creatingLobby = false
			m.joiningLobby = false
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

	if currentLobby == nil || m.joiningLobby {
		msg := "Connecting to server..."

		wsc := wsClient
		_ = wsc

		if wsClient.IsConnected() && m.creatingLobby {
			msg = "Loading lobby..."
		}

		if wsClient.IsConnected() && m.joiningLobby {
			msg = "Joining lobby..."
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
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s*", i+1, player.Name))
			continue
		}

		if player == nil {
			lobbyContent = append(lobbyContent, fmt.Sprintf("%d) ...", i+1))
			continue
		}

		lobbyContent = append(lobbyContent, fmt.Sprintf("%d) %s", i+1, player.Name))
	}

	lobbyContent = append(lobbyContent, fmt.Sprintf("Code: %s", currentLobby.Code))

	lobbyNamePane := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render(currentLobby.Name + "\n")

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
