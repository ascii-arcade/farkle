package lobby

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/client/eventloop"
	"github.com/ascii-arcade/farkle/client/game"
	"github.com/ascii-arcade/farkle/client/networkmanager"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/lobbies"
	"github.com/ascii-arcade/farkle/message"
	"github.com/ascii-arcade/farkle/player"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width  int
	height int

	lobby  *lobbies.Lobby
	player *player.Player

	errors         string
	networkManager *networkmanager.NetworkManager
}

type disconnectedMsg struct{}

func New(nm *networkmanager.NetworkManager) lobbyModel {
	return lobbyModel{
		networkManager: nm,
	}
}

func (m lobbyModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.networkManager.Outgoing <- message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeLeave,
				SentAt:  time.Now(),
			}
			return m, tea.Quit
		case "enter":
			if m.lobby != nil && m.lobby.Ready() {
				m.networkManager.Outgoing <- message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeStart,
					SentAt:  time.Now(),
					Data:    m.lobby.Code,
				}
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case eventloop.NetworkMsg:
		if msg.Data.Channel == message.ChannelPlayer {
			if msg.Data.Type == message.MessageTypeMe {
				if err := msg.Data.Unmarshal(&m.player); err != nil {
					return m, nil
				}
			}
		}

		if msg.Data.Channel == message.ChannelLobby {
			if err := msg.Data.Unmarshal(&m.lobby); err != nil {
				return m, nil
			}

			if m.lobby.Started {
				gameModel := game.NewModel(m.networkManager, m.lobby.Game, m.player)
				return gameModel, func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					}
				}
			}
		}
	}

	return m, nil
}

func (m lobbyModel) View() string {
	fullPaneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height-1).Align(lipgloss.Center, lipgloss.Center)
	lobbyStyle := lipgloss.NewStyle().Padding(1, 2).Margin(1).BorderStyle(lipgloss.NormalBorder()).Width(28)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width / 2)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.width / 2)

	if config.GetDebug() {
		fullPaneStyle = fullPaneStyle.BorderStyle(lipgloss.ASCIIBorder()).BorderForeground(lipgloss.Color("#0000ff")).Width(m.width - 2).Height(m.height - 3)
		controlsStyle = controlsStyle.Background(lipgloss.Color("#000066")).Foreground(lipgloss.Color("#ffffff"))
		errorsStyle = errorsStyle.Background(lipgloss.Color("#660000")).Foreground(lipgloss.Color("#ffffff"))
	}

	// if !me.Connected() {
	// 	return fullPaneStyle.Render(
	// 		lipgloss.JoinVertical(
	// 			lipgloss.Center,
	// 			lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render("Connecting..."),
	// 		),
	// 	)
	// }

	if m.lobby == nil {
		msg := ""
		// if m.creatingLobby {
		// 	msg = "Creating lobby..."
		// }

		// if m.joiningLobby {
		// 	msg = "Joining lobby..."
		// }

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

	playerList := []string{}

	for i, player := range m.lobby.Players {
		if player != nil && player.Host {
			playerList = append(playerList, fmt.Sprintf("%d) %s*", i+1, player.Name))
			continue
		}

		if player == nil {
			playerList = append(playerList, fmt.Sprintf("%d) ...", i+1))
			continue
		}

		playerList = append(playerList, fmt.Sprintf("%d) %s", i+1, player.Name))
	}

	lobbyNamePane := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render(fmt.Sprintf("Lobby Code: %s\n\n", m.lobby.Code))

	lobbyPane := lobbyStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lobbyNamePane,
			strings.Join(playerList, "\n"),
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
