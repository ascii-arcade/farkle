package game

import (
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/client/eventloop"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/game"
	"github.com/ascii-arcade/farkle/message"
	tea "github.com/charmbracelet/bubbletea"
)

func (m gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	gd := game.GameDetails{
		LobbyCode: m.game.LobbyCode,
		PlayerId:  m.player.Id,
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, nil
		}

		messageOut := message.Message{
			Channel: message.ChannelGame,
			SentAt:  time.Now(),
		}
		if msg.String() == "r" && !m.game.Rolled {
			messageOut.Type = message.MessageTypeRoll
		}

		if m.game.IsTurn(m.player) && m.game.Rolled {
			switch msg.String() {
			case "1", "2", "3", "4", "5", "6":
				gd.DieHeld, _ = strconv.Atoi(msg.String())
				messageOut.Type = message.MessageTypeHold
				messageOut.Data = gd.ToJSON()
			case "l":
				messageOut.Type = message.MessageTypeLock
			case "y":
				messageOut.Type = message.MessageTypeBank
			case "u":
				messageOut.Type = message.MessageTypeUndo
			}
		}
		if messageOut.Type != "" {
			messageOut.Data = gd.ToJSON()
			m.nm.Outgoing <- messageOut
		}
		return m, nil
	case rollMsg:
		if m.rollTickCount < rollFrames {
			m.rolling = true
			m.rollTickCount++
			m.poolRoll = dice.NewDicePool(6)
			m.poolRoll.Roll()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		m.rolling = false
		return m, nil
	case eventloop.NetworkMsg:
		if msg.Data.Channel == message.ChannelGame {
			switch msg.Data.Type {
			case message.MessageTypeUpdated:
				if err := msg.Data.Unmarshal(&m.game); err != nil {
					return m, nil
				}
			case message.MessageTypeRolled:
				if err := msg.Data.Unmarshal(&m.game); err != nil {
					return m, nil
				}

				m.rollTickCount = 0
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
