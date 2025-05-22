package game

import (
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/internal/client/eventloop"
	"github.com/ascii-arcade/farkle/internal/client/networkmanager"
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

	game   *game.Game
	player *player.Player
	nm     *networkmanager.NetworkManager
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

func NewModel(networkManager *networkmanager.NetworkManager, game *game.Game, player *player.Player) gameModel {
	return gameModel{
		poolRoll: dice.NewDicePool(6),
		game:     game,
		player:   player,
		nm:       networkManager,
	}
}

type rollMsg struct{}

func (m gameModel) Init() tea.Cmd {
	return tea.WindowSize()
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

		if m.game.IsTurn(m.player) {
			switch msg.String() {
			case "1", "2", "3", "4", "5", "6":
				choice, _ := strconv.Atoi(msg.String())
				gd := game.GameDetails{
					LobbyCode: m.game.LobbyCode,
					PlayerId:  m.player.Id,
					DieHeld:   choice,
				}

				m.nm.Outgoing <- message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeHold,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				}
				return m, nil
			case "r":
				gd := game.GameDetails{
					LobbyCode: m.game.LobbyCode,
					PlayerId:  m.player.Id,
				}

				m.nm.Outgoing <- message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeRoll,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				}
			case "l":
				gd := game.GameDetails{
					LobbyCode: m.game.LobbyCode,
					PlayerId:  m.player.Id,
				}
				m.nm.Outgoing <- message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeLock,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				}
			case "y":
				gd := game.GameDetails{
					LobbyCode: m.game.LobbyCode,
					PlayerId:  m.player.Id,
				}
				m.nm.Outgoing <- message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeBank,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				}
			case "u":
				gd := game.GameDetails{
					LobbyCode: m.game.LobbyCode,
					PlayerId:  m.player.Id,
				}
				m.nm.Outgoing <- message.Message{
					Channel: message.ChannelGame,
					Type:    message.MessageTypeUndo,
					SentAt:  time.Now(),
					Data:    gd.ToJSON(),
				}
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
		m.poolRoll = m.game.DicePool
		m.game.Log = append(m.game.Log, m.game.StyledPlayerName(m.game.Turn)+" rolled "+m.game.DicePool.RenderCharacters())
	case eventloop.NetworkMsg:
		if msg.Data.Channel == message.ChannelGame {
			switch msg.Data.Type {
			case message.MessageTypeUpdated:
				if err := msg.Data.Unmarshal(&m.game); err != nil {
					return m, nil
				}
			case message.MessageTypeRolled:
				m.rollTickCount = 0
				if err := msg.Data.Unmarshal(&m.game); err != nil {
					return m, nil
				}

				return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
					return rollMsg{}
				})
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, nil
}

func (m gameModel) View() string {
	cg := m.game
	_ = cg

	paneStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	poolPaneStyle := lipgloss.NewStyle().
		// Width(36).Height(10).
		Align(lipgloss.Center)

	logPaneStyle := lipgloss.NewStyle()

	if config.GetDebug() {
		paneStyle = paneStyle.
			Width(m.width - 1).
			Height(m.height - 1).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	poolRollPane := poolPaneStyle.Render(m.poolRoll.Render(0, 3) + "\n" + m.game.DicePool.Render(3, 6))
	poolHeldPane := poolPaneStyle.Render(m.game.DiceHeld.Render(0, 3) + "\n" + m.game.DiceHeld.Render(3, 6))

	centeredText := ""
	if m.error != "" {
		centeredText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	}

	poolPane := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolRollPane,
			poolHeldPane,
		),
		centeredText,
	)

	return paneStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			poolPane,
			m.game.PlayerScores(),
			logPaneStyle.Render(m.game.LogEntries()),
			"r to roll, l to lock, n to bust, y to bank, u to undo",
		),
	)
}
