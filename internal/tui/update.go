package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
			if m.poolRoll.contains(1) {
				m.poolRoll.remove(1)
				m.poolHeld.add(1)
			}
		case "2":
			if m.poolRoll.contains(2) {
				m.poolRoll.remove(2)
				m.poolHeld.add(2)
			}
		case "3":
			if m.poolRoll.contains(3) {
				m.poolRoll.remove(3)
				m.poolHeld.add(3)
			}
		case "4":
			if m.poolRoll.contains(4) {
				m.poolRoll.remove(4)
				m.poolHeld.add(4)
			}
		case "5":
			if m.poolRoll.contains(5) {
				m.poolRoll.remove(5)
				m.poolHeld.add(5)
			}
		case "6":
			if m.poolRoll.contains(6) {
				m.poolRoll.remove(6)
				m.poolHeld.add(6)
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
