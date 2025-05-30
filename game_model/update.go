package gamemodel

import (
	"slices"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width
		return m, nil
	case tea.KeyMsg:
		if !m.game.Started {

			switch msg.String() {
			case "s":
				if m.player.Host {
					m.game.Start()
				}
			}

			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		if m.game.IsTurn(m.player) {
			if msg.String() == "r" && !m.game.Rolled && !m.rolling {
				m.rollTickCount = 0
				m.rolling = true
				return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
					return rollMsg{}
				})
			}

			if m.game.Rolled && slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, msg.String()) {
				face, _ := strconv.Atoi(msg.String())
				if m.game.DicePool.Contains(face) {
					m.game.HoldDie(face)
					return m, nil
				}
			}

			switch msg.String() {
			case "l":
				_, err := m.game.DiceHeld.Score()
				if len(m.game.DiceHeld) != 0 && err == nil {
					m.game.LockDice()
				}
			case "y":
				if len(m.game.DiceLocked) >= 0 {
					m.game.Bank()
				}
			case "u":
				if len(m.game.DiceHeld) > 0 {
					m.game.Undo()
				}
			}
			return m, nil
		}
	case rollMsg:
		if m.rollTickCount < rollFrames {
			m.rollTickCount++
			m.game.DicePool.Roll()
			m.game.Refresh()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		m.game.RollDice()
		return m, nil
	case messages.RefreshGame:
		return m, waitForRefreshSignal(m.player.UpdateChan)
	}

	return m, nil
}
