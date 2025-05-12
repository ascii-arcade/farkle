package game

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/internal/config"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
	"github.com/ascii-arcade/farkle/internal/wsclient"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameModel struct {
	width  int
	height int

	error         string
	rollTickCount int
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

var (
	currentGame *Game
	me          *player.Player
	logger      *slog.Logger
	cmdSent     bool
)

func NewModel(logger *slog.Logger, p *player.Player, g *Game) gameModel {
	me = p
	currentGame = g
	logger = logger.With("component", "game")

	go func() {
		for msg := range wsclient.GameMessages {
			if msg.Type == message.MessageTypeUpdated {
				if err := json.Unmarshal([]byte(msg.Data.(string)), &currentGame); err != nil {
					continue
				}

				cmdSent = false
			}

			if msg.Type == message.MessageTypeRolled {
				// TODO: handle rolling
				if err := json.Unmarshal([]byte(msg.Data.(string)), &currentGame); err != nil {
					continue
				}

				cmdSent = false
			}
		}
	}()

	return gameModel{}
}

type tick struct{}
type rollMsg struct{}

func (m gameModel) Init() tea.Cmd {
	// go func() {
	// 	for {
	// 		if wsclient.GetClient() == nil {
	// 			logger.Debug("wsClient is nil, stopping monitoring for messages in lobby model")
	// 			return
	// 		}

	// 		select {
	// 		case <-wsclient.Disconnect:
	// 			logger.Debug("stopping monitoring for messages in lobby model")
	// 			return
	// 		case msg := <-wsclient.GameMessages:
	// 			switch msg.Type {
	// 			case message.MessageTypeUpdated:
	// 				logger.Debug("Received game update from server")
	// 				if err := json.Unmarshal([]byte(msg.Data.(string)), &currentGame); err != nil {
	// 					logger.Error("Error unmarshalling player message", "error", err)
	// 					continue
	// 				}
	// 			case message.MessageTypeRoll:
	// 				logger.Debug("Received game roll message from server")
	// 			}
	// 		}
	// 	}
	// }()
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tick{}
	})
}

func (m gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// case rollMsg:
	// 	if m.rollTickCount < rollFrames {
	// 		m.rollTickCount++
	// 		m.poolRoll.roll()
	// 		return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
	// 			return tickMsg{}
	// 		})
	// 	}
	// 	m.isRolling = false
	// 	m.log.add(m.styledPlayerName(m.currentPlayerIndex) + " rolled " + m.poolRoll.renderCharacters())
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
				gd := GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
					DieHeld:   choice,
				}

				wsclient.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeHold,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
			case "r":
				gd := GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				wsclient.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeRoll,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			case "l":
				gd := GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				wsclient.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeLock,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			case "y":
				gd := GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				wsclient.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeBank,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			case "u":
				gd := GameDetails{
					LobbyCode: currentGame.LobbyCode,
					PlayerId:  me.Id,
				}
				wsclient.SendMessage(message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeUndo,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				})
				cmdSent = true
			}
		}
	case struct{}:
		return m, nil
	case tick:
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			cg := currentGame
			_ = cg
			return tick{}
		})
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

	poolRollPane := poolPaneStyle.Render(currentGame.DicePool.Render(0, 3) + "\n" + currentGame.DicePool.Render(3, 6))
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
