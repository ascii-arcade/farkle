package game

import (
	"strconv"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/score"
	"github.com/charmbracelet/lipgloss"
)

func (m gameModel) View() string {
	paneStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	poolPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		PaddingBottom(1).
		Width(31).
		Height(10)
	heldPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Width(32).
		Height(11)
	heldScorePaneStyle := lipgloss.NewStyle()

	logPaneStyle := lipgloss.NewStyle()

	if config.GetDebug() {
		paneStyle = paneStyle.
			Width(m.width - 2).
			Height(m.height - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	poolDie := ""
	if m.rolling {
		poolDie = m.poolRoll.Render(0, 3) + "\n" + m.poolRoll.Render(3, 6)
	} else {
		poolDie = m.game.DicePool.Render(0, 3) + "\n" + m.game.DicePool.Render(3, 6)
	}

	poolRollPane := lipgloss.JoinVertical(
		lipgloss.Left,
		poolPaneStyle.Render(poolDie),
	)

	heldScore, ok := score.Calculate(m.game.DiceHeld)
	if !ok {
		heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color(colorError))
	}

	heldDie := lipgloss.NewStyle().
		Height(10).
		Render(m.game.DiceHeld.Render(0, 3) + "\n" + m.game.DiceHeld.Render(3, 6))

	poolHeldPane := heldPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		heldDie,
		heldScorePaneStyle.Render("Score: "+strconv.Itoa(heldScore)),
	))

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
