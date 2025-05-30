package gamemodel

import (
	"strconv"
	"strings"

	"github.com/ascii-arcade/farkle/config"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	paneStyle := m.style.
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)
	logPaneStyle := m.style.
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Height(12).
		Width(35)
	poolPaneStyle := m.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Padding(1, 0).
		Width(32).
		Height(12)
	heldPaneStyle := m.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Width(48).
		Height(12)
	heldScorePaneStyle := m.style
	lockedPaneStyle := m.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Padding(0, 1).
		Width(19).
		Height(12)

	if !m.game.Started {
		playersString := []string{}
		for _, player := range m.game.GetPlayers() {
			n := player.Name
			if player.Host {
				n += " (host)"
			}
			if player.Name == m.player.Name {
				n += " (you)"
			}

			playersString = append(playersString, n)
		}

		lobbyPaneStyle := m.style.
			Align(lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6")).
			Height(12).
			Width(40)

		lobbyPane := lobbyPaneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				m.style.Render(
					lipgloss.JoinVertical(
						lipgloss.Center,
						[]string{
							"Game Code: " + m.game.Code + "\n",
							strings.Join(playersString, "\n"),
						}...,
					),
				),
			),
		)
		out := []string{
			lobbyPane,
		}
		if m.player.Host {
			out = append(out, m.style.Render("Press 's' to start the game"))
		} else {
			out = append(out, m.style.Render("Waiting for host to start the game..."))
		}
		return paneStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			out...,
		))
	}

	if config.GetDebug() {
		paneStyle = paneStyle.
			Width(m.width - 2).
			Height(m.height - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	poolDie := m.game.DicePool.Render(0, 6)

	poolRollPane := lipgloss.JoinVertical(
		lipgloss.Left,
		poolPaneStyle.Render(poolDie),
	)

	heldScore, err := m.game.DiceHeld.Score()
	if err != nil {
		m.error = err.Error()
		heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color(colorError))
	}

	heldDie := m.style.
		Height(10).
		Render(m.game.DiceHeld.Render(0, 6))

	if m.game.Busted {
		heldDie = m.style.
			Height(10).
			Foreground(lipgloss.Color(colorError)).
			Align(lipgloss.Center, lipgloss.Center).
			Render("BUSTED")
		heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color("#ff0000"))
	}

	poolHeldPane := heldPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"To be Locked (l)",
		heldDie,
		heldScorePaneStyle.Render("Score: "+strconv.Itoa(heldScore)),
	))

	bankedDie := ""
	for _, diePool := range m.game.DiceLocked {
		bankedDie += diePool.RenderCharacters() + "\n"
	}
	lockedScore := 0
	for _, diePool := range m.game.DiceLocked {
		ls, err := diePool.Score()
		if err != nil {
			m.error = err.Error()
		} else {
			lockedScore += ls
		}
	}
	lockedPane := lockedPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"To be Banked (y)",
		m.style.
			Height(10).Render(bankedDie),
		"Score: "+strconv.Itoa(lockedScore),
	))

	centeredText := ""
	if m.error != "" {
		centeredText = m.style.Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	}

	poolPane := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			logPaneStyle.Render(m.game.RenderLog(12)),
			poolRollPane,
		),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolHeldPane,
			lockedPane,
		),
		centeredText,
	)

	controls := "r to roll, l to lock, y to bank, u to undo, esc to quit"
	if m.player.Host {
		controls += ", ctrl+r to reset"
	}

	return paneStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			poolPane,
			m.game.PlayerScores(),
			controls,
		),
	)
}
