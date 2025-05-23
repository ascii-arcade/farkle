package game

import (
	"github.com/ascii-arcade/farkle/config"
	"github.com/charmbracelet/lipgloss"
)

func (m gameModel) View() string {
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
			Width(m.width - 2).
			Height(m.height - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	poolRollPane := ""
	if m.rolling {
		poolRollPane = poolPaneStyle.Render(m.poolRoll.Render(0, 3) + "\n" + m.poolRoll.Render(3, 6))
	} else {
		poolRollPane = m.game.DicePool.Render(0, 3) + "\n" + m.game.DicePool.Render(3, 6)
	}
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
			lipgloss.NewStyle().MarginLeft(5).Render(poolHeldPane),
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
