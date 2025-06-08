package board

import (
	"fmt"
	"strings"

	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/charmbracelet/lipgloss"
)

func (s *tableScreen) View() string {
	playerColor := s.model.game.GetTurnPlayer().Color

	paneStyle := s.model.style.
		Width(s.model.width-2).
		Height(s.model.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.model.player.Color)
	logPaneStyle := s.model.style.
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(playerColor).
		Height(12).
		Width(35)
	poolPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(playerColor).
		Padding(1, 0).
		Width(32).
		Height(12)
	heldPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(playerColor).
		Width(48).
		Height(12)
	heldScorePaneStyle := s.model.style
	lockedPaneStyle := s.model.style.
		Align(lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(playerColor).
		Padding(0, 1).
		Width(19).
		Height(12)

	if !s.model.game.InProgress {
		playerNames := []string{}
		for _, player := range s.model.game.GetPlayers() {
			n := player.StyledPlayerName(s.model.style)
			if player.IsHost {
				n += fmt.Sprintf(" (%s)", s.model.lang().Get("board", "player_list_host"))
			}
			if player.Name == s.model.player.Name {
				n += fmt.Sprintf(" (%s)", s.model.lang().Get("board", "player_list_you"))
			}

			playerNames = append(playerNames, n)
		}

		lobbyPaneStyle := s.model.style.
			Align(lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Farkle).
			Height(12).
			Width(40)

		lobbyPane := lobbyPaneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				s.model.style.Render(
					lipgloss.JoinVertical(
						lipgloss.Center,
						[]string{
							fmt.Sprintf("%s: %s\n", s.model.lang().Get("board", "game_code"), s.model.game.Code),
							strings.Join(playerNames, "\n"),
						}...,
					),
				),
			),
		)

		var statusMsg string
		switch {
		case s.model.player.IsHost && s.model.game.Ready():
			statusMsg = fmt.Sprintf(s.model.lang().Get("board", "press_to_start"), keys.LobbyStartGame.String(s.model.style))
		case s.model.player.IsHost:
			statusMsg = s.model.lang().Get("board", "waiting_for_players")
		default:
			statusMsg = s.model.lang().Get("board", "waiting_for_start")
		}

		return paneStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			lobbyPane,
			statusMsg,
		))
	}

	if s.model.game.IsGameOver() {
		winner := s.model.game.GetWinningPlayer()
		return paneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				s.model.style.Bold(true).Foreground(colors.Farkle).Render(s.model.lang().Get("board", "game_over")),
				s.model.style.Bold(true).Render(fmt.Sprintf(s.model.lang().Get("board", "winner"), winner.StyledPlayerName(s.model.style))),
				s.model.style.Render(fmt.Sprintf(s.model.lang().Get("board", "host_can_restart"), keys.RestartGame.String(s.model.style))),
			),
		)
	}

	if dcPlayers := s.model.game.GetDisconnectedPlayers(); len(dcPlayers) > 0 {
		return paneStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				s.model.game.Code,
				s.model.style.Bold(true).Render(fmt.Sprintf(s.model.lang().Get("error", "disconnected"), dcPlayers[0].StyledPlayerName(s.model.style))),
			),
		)
	}

	poolRollStrings := []string{}
	if s.model.game.GetTurnPlayer().Name == s.model.player.Name {
		poolPaneStyle = poolPaneStyle.Padding(0, 0, 1, 0)
		poolRollStrings = append(poolRollStrings, s.model.lang().Get("board", "your_turn")+"\n")
	}
	poolRollStrings = append(poolRollStrings, s.model.game.DicePool.Render(0, 6))
	poolRollPane := lipgloss.JoinVertical(
		lipgloss.Left,
		poolPaneStyle.Render(poolRollStrings...),
	)

	heldScore, _, err := s.model.game.DiceHeld.Score()
	if err != nil {
		s.model.error = err.Error()
		heldScorePaneStyle = heldScorePaneStyle.Foreground(colors.Error)
	}

	heldDie := s.model.style.
		Height(10).
		Render(s.model.game.DiceHeld.Render(0, 6))

	if s.model.game.Busted {
		heldDie = s.model.style.
			Height(10).
			Foreground(colors.Error).
			Align(lipgloss.Center, lipgloss.Center).
			Render(s.model.lang().Get("board", "busted"))
		heldScorePaneStyle = heldScorePaneStyle.Foreground(colors.Error)
	}

	poolHeldPane := heldPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf(s.model.lang().Get("board", "to_be_locked"), keys.ActionLock.String(s.model.style)),
		heldDie,
		heldScorePaneStyle.Render(fmt.Sprintf(s.model.lang().Get("board", "score"), heldScore)),
	))

	bankedDie := ""
	for _, diePool := range s.model.game.DiceLocked {
		bankedDie += diePool.RenderCharacters() + "\n"
	}
	lockedScore := 0
	for _, diePool := range s.model.game.DiceLocked {
		ls, _, err := diePool.Score()
		if err != nil {
			s.model.error = err.Error()
		} else {
			lockedScore += ls
		}
	}
	lockedPane := lockedPaneStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf(s.model.lang().Get("board", "to_be_banked"), keys.ActionBank.String(s.model.style)),
		s.model.style.Height(10).Render(bankedDie),
		fmt.Sprintf(s.model.lang().Get("board", "score"), lockedScore),
	))

	centeredText := ""
	if s.model.error != "" {
		centeredText = s.model.style.Bold(true).Foreground(colors.Error).Render(s.model.error)
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

	controls := strings.Join([]string{
		fmt.Sprintf(s.model.lang().Get("board", "controls", "roll"), keys.ActionRoll.String(s.model.style)),
		fmt.Sprintf(s.model.lang().Get("board", "controls", "lock"), keys.ActionLock.String(s.model.style)),
		fmt.Sprintf(s.model.lang().Get("board", "controls", "bank"), keys.ActionBank.String(s.model.style)),
		fmt.Sprintf(s.model.lang().Get("board", "controls", "undo"), keys.ActionUndo.String(s.model.style)),
		fmt.Sprintf(s.model.lang().Get("board", "controls", "help"), keys.OpenHelp.String(s.model.style)),
		fmt.Sprintf(s.model.lang().Get("global", "quit"), keys.ExitApplication.String(s.model.style)),
	}, ", ")

	return paneStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			poolPane,
			s.model.game.PlayerScores(),
			controls,
		),
	)
}
