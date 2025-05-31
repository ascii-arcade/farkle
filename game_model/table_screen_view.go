package gamemodel

import (
	"strconv"
	"strings"

	"github.com/ascii-arcade/farkle/config"
	"github.com/charmbracelet/lipgloss"
)

type tableScreen struct {
	model *Model
}

func (s *tableScreen) setModel(model *Model) {
	s.model = model
}

func (s *tableScreen) view() string {
	borderColor := s.model.game.GetTurnPlayer().Color

	paneStyle := s.model.style.
		Width(s.model.width).
		Height(s.model.height).
		Align(lipgloss.Center, lipgloss.Center)
	logPaneStyle := s.model.style.
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Height(12).
		Width(35)
	poolPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 0).
		Width(32).
		Height(12)
	heldPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(48).
		Height(12)
	heldScorePaneStyle := s.model.style
	lockedPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(19).
		Height(12)

	if !s.model.game.Started {
		playersString := []string{}
		for _, player := range s.model.game.GetPlayers() {
			n := player.Name
			if player.Host {
				n += " (host)"
			}
			if player.Name == s.model.player.Name {
				n += " (you)"
			}

			playersString = append(playersString, n)
		}

		lobbyPaneStyle := s.model.style.
			Align(lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6")).
			Height(12).
			Width(40)

		lobbyPane := lobbyPaneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				s.model.style.Render(
					lipgloss.JoinVertical(
						lipgloss.Center,
						[]string{
							"Game Code: " + s.model.game.Code + "\n",
							strings.Join(playersString, "\n"),
						}...,
					),
				),
			),
		)
		out := []string{
			lobbyPane,
		}
		if s.model.player.Host {
			out = append(out, s.model.style.Render("Press 's' to start the game"))
		} else {
			out = append(out, s.model.style.Render("Waiting for host to start the game..."))
		}
		return paneStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			out...,
		))
	}

	if m.game.IsGameOver() {
		winner := m.game.GetWinningPlayer()
		return paneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				m.style.Bold(true).Foreground(lipgloss.Color("#3B82F6")).Render("Game Over!"),
				m.style.Bold(true).Render("Winner: "+winner.StyledPlayerName(m.style)),
				m.style.Render("The host can press 'r' to restart the game"),
			),
		)
	}

	if config.GetDebug() {
		paneStyle = paneStyle.
			Width(s.model.width - 2).
			Height(s.model.height - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(lipgloss.Color("#ff0000"))
	}

	poolRollStrings := []string{}
	if s.model.game.GetTurnPlayer().Id == s.model.player.Id {
		poolPaneStyle = poolPaneStyle.Padding(0, 0, 1, 0)
		poolRollStrings = append(poolRollStrings, "Your Turn!\n")
	}
	poolRollStrings = append(poolRollStrings, s.model.game.DicePool.Render(0, 6))
	poolRollPane := lipgloss.JoinVertical(
		lipgloss.Left,
		poolPaneStyle.Render(poolRollStrings...),
	)

	heldScore, err := s.model.game.DiceHeld.Score()
	if err != nil {
		s.model.error = err.Error()
		heldScorePaneStyle = heldScorePaneStyle.Foreground(lipgloss.Color(colorError))
	}

	heldDie := s.model.style.
		Height(10).
		Render(s.model.game.DiceHeld.Render(0, 6))

	if s.model.game.Busted {
		heldDie = s.model.style.
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
	for _, diePool := range s.model.game.DiceLocked {
		bankedDie += diePool.RenderCharacters() + "\n"
	}
	lockedScore := 0
	for _, diePool := range s.model.game.DiceLocked {
		ls, err := diePool.Score()
		if err != nil {
			s.model.error = err.Error()
		} else {
			lockedScore += ls
		}
	}
	lockedPane := lockedPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"To be Banked (y)",
		s.model.style.
			Height(10).Render(bankedDie),
		"Score: "+strconv.Itoa(lockedScore),
	))

	centeredText := ""
	if s.model.error != "" {
		centeredText = s.model.style.Bold(true).Foreground(lipgloss.Color(colorError)).Render(s.model.error)
	}

	poolPane := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			logPaneStyle.Render(s.model.game.RenderLog(12)),
			poolRollPane,
		),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolHeldPane,
			lockedPane,
		),
		centeredText,
	)

	controls := "r to roll, l to lock, y to bank, u to undo, ? for help, esc to quit"

	return paneStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			poolPane,
			s.model.game.PlayerScores(),
			controls,
		),
	)
}
