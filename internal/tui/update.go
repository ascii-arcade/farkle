package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthibodeaux/go-farkle/internal/score"
)

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.isRolling {
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.isRolling = true
			m.tickCount = 0
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		case "1":
			m.handleNumber(1)
		case "2":
			m.handleNumber(2)
		case "3":
			m.handleNumber(3)
		case "4":
			m.handleNumber(4)
		case "5":
			m.handleNumber(5)
		case "6":
			m.handleNumber(6)
		case "l":
			if len(m.poolHeld) > 0 {
				score, err := score.Calculate(m.poolHeld)
				if err == nil {
					m.lockedInScore += score
					m.poolHeld = newDicePool(0)
				}
				if len(m.poolRoll) == 0 {
					m.poolRoll = newDicePool(6)
				}
			}
		case "u":
			if len(m.poolHeld) > 0 {
				die := m.poolHeld[len(m.poolHeld)-1]
				m.poolRoll.add(die)
				m.poolHeld.remove(die)
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		if m.tickCount < rollFrames {
			m.tickCount++
			m.poolRoll.roll()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		}
		m.isRolling = false

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *model) handleNumber(n int) {
	if m.poolRoll.contains(n) {
		m.poolRoll.remove(n)
		m.poolHeld.add(n)
	}
}
