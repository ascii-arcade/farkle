package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/config"
	"github.com/ascii-arcade/farkle/internal/game"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyModel struct {
	width  int
	height int

	creatingLobby bool
	joiningLobby  bool

	errors    string
	menuModel menuModel
}

type disconnectedMsg struct{}

func newLobbyModel(playerName string, code string, joining bool) (lobbyModel, tea.Cmd) {
	messages = make(chan message.Message, 100)
	gameMessages = make(chan message.Message, 100)

	me = player.NewPlayer(logger, nil, playerName)
	me.Connect(code, messages)

	m := lobbyModel{
		joiningLobby: joining,
	}

	return m, m.Init()
}

func (m lobbyModel) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), watchForMessages())
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			_ = me.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeLeave,
				SentAt:  time.Now(),
			})

			currentLobby = nil

			return newMenu(), m.menuModel.Init()
		case "enter":
			if currentLobby != nil && currentLobby.Ready() {
				_ = me.SendMessage(message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeStart,
					SentAt:  time.Now(),
					Data:    currentLobby.Code,
				})
			}

			return m, nil
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case message.Message:
		logger.Debug("Received message from server", "channel", msg.Channel, "type", msg.Type)
		switch msg.Channel {
		case message.ChannelPing:
		case message.ChannelLobby:
			switch msg.Type {
			case message.MessageTypeUpdated:
				logger.Debug("Received lobby update from server")
				if err := msg.Unmarshal(&currentLobby); err != nil {
					logger.Error("Error unmarshalling player message", "error", err)
					break
				}

				if currentLobby.Started && currentLobby.Game != nil {
					_, _ = tea.NewProgram(game.NewModel(logger, me, currentLobby.Game), tea.WithAltScreen()).Run()
				}
			}
		default:
		}
		return m, watchForMessages()
	case disconnectedMsg:
		logger.Debug("stopping monitoring for messages in lobby model")
		return m, tea.Quit
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

	if !me.Connected() {
		return fullPaneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render("Connecting..."),
			),
		)
	}

	if currentLobby == nil {
		msg := ""
		if m.creatingLobby {
			msg = "Creating lobby..."
		}

		if m.joiningLobby {
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

	playerList := []string{}

	for i, player := range currentLobby.Players {
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

	lobbyNamePane := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Render(fmt.Sprintf("Lobby Code: %s\n\n", currentLobby.Code))

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

func watchForMessages() tea.Cmd {
	return func() tea.Msg {
		for {
			if me == nil {
				logger.Debug("me is nil, stopping monitoring for messages in lobby model")
				return disconnectedMsg{}
			}

			select {
			case <-me.Disconnected():
				logger.Debug("stopping monitoring for messages in lobby model")
				return disconnectedMsg{}
			case msg := <-messages:
				logger.Debug("Received message from server", "channel", msg.Channel, "type", msg.Type)
				return msg
			}
		}
	}
}
