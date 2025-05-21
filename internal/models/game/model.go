package game

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/internal/config"
	"github.com/ascii-arcade/farkle/internal/dice"
	"github.com/ascii-arcade/farkle/internal/game"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameModel struct {
	width         int
	height        int
	isRolling     bool
	poolRoll      dice.DicePool
	error         string
	rollTickCount int

	game *game.Game
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

var (
	currentGame *game.Game
	me          *player.Player
	logger      *slog.Logger
	cmdSent     bool
	messages    chan message.Message
)

func NewModel(loggerIn *slog.Logger, p *player.Player, g *game.Game) gameModel {
	messages = make(chan message.Message, 100)
	me = p
	currentGame = g
	logger = loggerIn.With("component", "game")
	go me.MonitorGameMessages(messages)

	return gameModel{
		poolRoll: dice.NewDicePool(6),
	}
}

type disconnectedMsg struct{}
type rollMsg struct{}

func (m gameModel) Init() tea.Cmd {
	return watchForMessages()
}

func (m gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, nil

		}

		if currentGame.IsTurn(me) {
			switch msg.String() {
			case "1", "2", "3", "4", "5", "6":
				choice, _ := strconv.Atoi(msg.String())
				gd := game.GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
					DieHeld:   choice,
				}

				me.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeHold,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
			case "r":
				gd := game.GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				me.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeRoll,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
				return m, tea.Cmd(func() tea.Msg {
					<-currentGame.roll
					return rollMsg{}
				})
			case "l":
				gd := game.GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				me.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeLock,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			case "y":
				gd := game.GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				me.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeBank,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			case "u":
				gd := game.GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				me.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeUndo,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			}
		}
	case rollMsg:
		if m.rollTickCount < rollFrames {
			m.rollTickCount++
			m.poolRoll.Roll()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		m.isRolling = false
		m.poolRoll = currentGame.DicePool
		currentGame.log = append(currentGame.log, currentGame.styledPlayerName(currentGame.Turn)+" rolled "+currentGame.DicePool.RenderCharacters())
	case message.Message:
		logger.Debug("Received message from server", "channel", msg.Channel, "type", msg.Type)
		switch msg.Type {
		case message.MessageTypeUpdated:
			logger.Debug("Received game update from server")
			if err := msg.Unmarshal(&currentGame); err != nil {
				logger.Error("Error unmarshalling player message", "error", err)
				return m, nil
			}
		}
		return m, watchForMessages()
	case disconnectedMsg:
		logger.Debug("stopping monitoring for messages in lobby model")
		return m, tea.Quit
	default:
	}

	return m, nil
}

func (m gameModel) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	cg := currentGame
	_ = cg

	debugPaneStyle := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Left)
	poolPaneStyle := lipgloss.NewStyle().Width(36).Height(10).Align(lipgloss.Center)

	poolRollPane := poolPaneStyle.Render(m.poolRoll.Render(0, 3) + "\n" + currentGame.DicePool.Render(3, 6))
	poolHeldPane := poolPaneStyle.Render(currentGame.DiceHeld.Render(0, 3) + "\n" + currentGame.DiceHeld.Render(3, 6))

	centeredText := ""
	if m.error != "" {
		centeredText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	}

	debugPane := ""

	if config.GetDebug() {
		debugMsgs := []string{
			"Debug",
			"Current Player: " + currentGame.Players[currentGame.Turn].Name,
		}
		debugPane = lipgloss.JoinHorizontal(
			lipgloss.Left,
			debugPaneStyle.Render(debugMsgs...),
		)
	}

	poolPanes := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolRollPane,
			poolHeldPane,
		),
		centeredText,
	)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		"r to roll, l to lock, n to bust, y to bank, u to undo",
		lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			poolPanes,
			currentGame.playerScores(),
			"",
			m.logPane(),
			debugPane,
		),
	)

	return style.Render(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			panes,
		),
	)
}

func (m *gameModel) logPane() string {
	return lipgloss.NewStyle().Width(80).Height(15).Render(currentGame.log.entries())
}

func watchForMessages() tea.Cmd {
	return func() tea.Msg {
		for {
			if currentGame == nil {
				logger.Debug("currentGame is nil, stopping monitoring for messages in lobby model")
				return disconnectedMsg{}
			}

			select {
			case <-me.Disconnected():
				logger.Debug("stopping monitoring for messages in game model")
				return disconnectedMsg{}
			case msg := <-messages:
				logger.Debug("Received game message from server", "channel", msg.Channel, "type", msg.Type)
				return msg
			}
		}
	}
}
