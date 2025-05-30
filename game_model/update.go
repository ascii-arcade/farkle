package gamemodel

import (
	"github.com/ascii-arcade/farkle/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// gd := GameDetails{
	// 	LobbyCode: m.game.LobbyCode,
	// 	PlayerId:  m.player.Id,
	// }
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		// 	case tea.KeyCtrlR:
		// 		if m.player.Host {
		// 			m.nm.Outgoing <- message.Message{
		// 				Channel: message.ChannelGame,
		// 				SentAt:  time.Now(),
		// 				Type:    message.MessageTypeStart,
		// 				Data:    gd.ToJSON(),
		// 			}
		// 		}
		// 		return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		// 	if m.game.IsTurn(m.player) {
		// 		messageOut := message.Message{
		// 			Channel: message.ChannelGame,
		// 			SentAt:  time.Now(),
		// 		}

		// 		if msg.String() == "r" && !m.game.Rolled {
		// 			messageOut.Type = message.MessageTypeRoll
		// 		}

		// 		if m.game.Rolled && slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, msg.String()) {
		// 			face, _ := strconv.Atoi(msg.String())
		// 			if !m.game.DicePool.Contains(face) {
		// 				return m, nil
		// 			}
		// 			gd.DieHeld, _ = strconv.Atoi(msg.String())
		// 			messageOut.Type = message.MessageTypeHold
		// 			messageOut.Data = gd.ToJSON()
		// 		}

		// 		switch msg.String() {
		// 		case "l":
		// 			if len(m.game.DiceHeld) == 0 {
		// 				return m, nil
		// 			}
		// 			messageOut.Type = message.MessageTypeLock
		// 		case "y":
		// 			if len(m.game.DiceLocked) == 0 {
		// 				return m, nil
		// 			}
		// 			messageOut.Type = message.MessageTypeBank
		// 		case "u":
		// 			if len(m.game.DiceHeld) == 0 {
		// 				return m, nil
		// 			}
		// 			messageOut.Type = message.MessageTypeUndo
		// 		}

		// 		if messageOut.Type != "" {
		// 			messageOut.Data = gd.ToJSON()
		// 			m.nm.Outgoing <- messageOut
		// 		}
		// 		return m, nil
		// 	}
		// case rollMsg:
		// 	if m.rollTickCount < rollFrames {
		// 		m.rolling = true
		// 		m.rollTickCount++
		// 		m.poolRoll = dice.NewDicePool(len(m.game.DicePool))
		// 		m.poolRoll.Roll()
		// 		return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
		// 			return rollMsg{}
		// 		})
		// 	}
		// 	m.rolling = false
		// 	return m, nil
		// case eventloop.NetworkMsg:
		// 	if msg.Data.Channel == message.ChannelGame {
		// 		switch msg.Data.Type {
		// 		case message.MessageTypeUpdated:
		// 			if err := msg.Data.Unmarshal(&m.game); err != nil {
		// 				return m, nil
		// 			}
		// 		case message.MessageTypeRolled:
		// 			if err := msg.Data.Unmarshal(&m.game); err != nil {
		// 				return m, nil
		// 			}

		// 			m.rollTickCount = 0
		// 			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
		// 				return rollMsg{}
		// 			})
		// 		}
		// 	}
	case messages.RefreshGame:
		return m, waitForRefreshSignal(m.player.UpdateChan)
	}

	return m, nil
}
