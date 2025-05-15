package menu

import (
	"encoding/json"
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

	started bool

	errors    string
	menuModel menuModel
}

func newLobbyModel(playerName string, code string, joining bool) (lobbyModel, tea.Cmd) {
	messages = make(chan message.Message, 100)

	me = player.NewPlayer(logger, nil, playerName)
	me.Connect(code, messages)

	m := lobbyModel{
		joiningLobby: joining,
	}

	return m, m.Init()
}

func (m lobbyModel) Init() tea.Cmd {
	go func() {
		for {
			if me == nil {
				logger.Debug("me is nil, stopping monitoring for messages in lobby model")
				return
			}

			select {
			case <-me.Disconnected():
				logger.Debug("stopping monitoring for messages in lobby model")
				return
			case msg := <-messages:
				switch msg.Type {
				case message.MessageTypeMe:
					logger.Debug("Received player message from server")
					if err := json.Unmarshal([]byte(msg.Data), &me); err != nil {
						logger.Error("Error unmarshalling player message", "error", err)
						continue
					}
				case message.MessageTypeUpdated:
					logger.Debug("Received lobby update from server")
					if err := msg.Unmarshal(&currentLobby); err != nil {
						logger.Error("Error unmarshalling player message", "error", err)
						continue
					}

					if currentLobby.Started && currentLobby.Game != nil {
						tea.NewProgram(game.NewModel(logger, me, currentLobby.Game), tea.WithAltScreen()).Run()
					}
				}
			}
		}
	}()
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
			me.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeLeave,
				SentAt:  time.Now(),
			})

			currentLobby = nil

			return newMenu(), m.menuModel.Init()
		case "enter":
			if currentLobby.Ready() {
				// 	tui.RunFromLobby(currentLobby)
				// 	return m.menuModel, nil
				me.SendMessage(message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeStart,
					SentAt:  time.Now(),
					Data:    currentLobby.Code,
				})
			}

			// m.errors = "Please wait for at least two players to join before starting the game"

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
