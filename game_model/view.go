package gamemodel

import (
	"github.com/ascii-arcade/farkle/config"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	paneStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center)
	poolPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Padding(1, 0).
		Width(32).
		Height(12)
	heldPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Width(31).
		Height(12)
	heldScorePaneStyle := lipgloss.NewStyle()
	lockedPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Padding(0, 1).
		Width(19).
		Height(12)
	_ = lockedPaneStyle
	_ = heldScorePaneStyle
	_ = heldPaneStyle
	_ = poolPaneStyle

	logPaneStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Height(12).
		Width(35)
	_ = logPaneStyle

	if config.GetDebug() {
		paneStyle = paneStyle.
			Width(m.Width - 2).
			Height(m.Height - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	return paneStyle.Render("")

	// poolDie := m.game.DicePool.Render(0, 6)
	// if m.rolling {
	// 	poolDie = m.poolRoll.Render(0, 6)
	// }

	// poolRollPane := lipgloss.JoinVertical(
	// 	lipgloss.Left,
	// 	poolPaneStyle.Render(poolDie),
	// )

	// heldScore, ok := score.Calculate(m.game.DiceHeld)
	// if !ok {
	// 	heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color(colorError))
	// }

	// heldDie := lipgloss.NewStyle().
	// 	Height(10).
	// 	Render(m.game.DiceHeld.Render(0, 6))

	// if m.game.Busted {
	// 	heldDie = lipgloss.NewStyle().
	// 		Height(10).
	// 		Foreground(lipgloss.Color(colorError)).
	// 		Align(lipgloss.Center, lipgloss.Center).
	// 		Render("BUSTED")
	// 	heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color("#ff0000"))
	// }

	// poolHeldPane := heldPaneStyle.Render(lipgloss.JoinVertical(
	// 	lipgloss.Left,
	// 	"To be Locked (l)",
	// 	heldDie,
	// 	heldScorePaneStyle.Render("Score: "+strconv.Itoa(heldScore)),
	// ))

	// bankedDie := ""
	// for _, diePool := range m.game.DiceLocked {
	// 	bankedDie += diePool.RenderCharacters() + "\n"
	// }
	// lockedScore := 0
	// for _, diePool := range m.game.DiceLocked {
	// 	score, _ := diePool.Score()
	// 	lockedScore += score
	// }
	// lockedPane := lockedPaneStyle.Render(lipgloss.JoinVertical(
	// 	lipgloss.Left,
	// 	"To be Banked (y)",
	// 	lipgloss.NewStyle().
	// 		Height(10).Render(bankedDie),
	// 	"Score: "+strconv.Itoa(lockedScore),
	// ))

	// centeredText := ""
	// if m.error != "" {
	// 	centeredText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	// }

	// poolPane := lipgloss.JoinVertical(
	// 	lipgloss.Center,
	// 	lipgloss.JoinHorizontal(
	// 		lipgloss.Top,
	// 		logPaneStyle.Render(m.game.RenderLog(12)),
	// 		poolRollPane,
	// 		poolHeldPane,
	// 		lockedPane,
	// 	),
	// 	centeredText,
	// )

	// controls := "r to roll, l to lock, y to bank, u to undo, esc to quit"
	// if m.player.Host {
	// 	controls += ", ctrl+r to reset"
	// }

	// return paneStyle.Render(
	// 	lipgloss.JoinVertical(
	// 		lipgloss.Left,
	// 		poolPane,
	// 		m.game.PlayerScores(),
	// 		controls,
	// 	),
	// )
}
