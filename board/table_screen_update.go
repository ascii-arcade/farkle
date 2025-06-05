package board

import (
	"slices"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/score"
	tea "github.com/charmbracelet/bubbletea"
)

func (s *tableScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	if !s.model.game.IsTurn(s.model.player) {
		return s.model, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case rollMsg:
		if s.rollTickCount < rollFrames {
			s.rollTickCount++
			s.model.game.DicePool.Roll()
			s.model.game.Refresh()
			return s.model, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		s.rolling = false
		s.model.game.RollDice()

	case tea.KeyMsg:
		s.model.error = ""
		switch msg.String() {
		case "r":
			if s.model.game.IsGameOver() && s.model.player.IsHost {
				s.model.game.Restart()
				return s.model, nil
			}

			if !s.model.game.Rolled && !s.rolling {
				s.rollTickCount = 0
				s.rolling = true
				return s.model, tea.Tick(rollInterval, func(time.Time) tea.Msg {
					return rollMsg{}
				})
			}

		case "1", "2", "3", "4", "5", "6":
			if s.model.game.Rolled && slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, msg.String()) {
				face, _ := strconv.Atoi(msg.String())
				if s.model.game.DicePool.Contains(face) {
					s.model.game.HoldDie(face)
					return s.model, nil
				}
			}

		case "!", "@", "#", "$", "%", "^":
			if s.model.game.Rolled {
				faceMap := map[string]int{"!": 1, "@": 2, "#": 3, "$": 4, "%": 5, "^": 6}
				face := faceMap[msg.String()]
				c := 0
				for _, die := range s.model.game.DicePool {
					if die == face {
						c++
					}
				}
				for range c {
					s.model.game.HoldDie(face)
				}

			}

		case "l":
			_, err := s.model.game.DiceHeld.Score()
			if len(s.model.game.DiceHeld) != 0 && err == nil {
				if err := s.model.game.LockDice(); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}
			}

		case "y", "b":
			if len(s.model.game.DiceHeld) == 0 && len(s.model.game.DiceLocked) > 0 {
				if err := s.model.game.Bank(); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}
			}

		case "a":
			for _, face := range score.GetScorableDieFaces(s.model.game.DicePool) {
				s.model.game.HoldDie(face)
			}

		case "u", "backspace":
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.Undo()
			}

		case "U":
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.UndoAll()
			}

		case "c":
			if err := s.model.game.ClearHeld(); err != nil {
				s.model.error = s.model.lang().Get("error", "game", err.Error())
				return s.model, nil
			}

		case "?":
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: &helpScreen{model: s.model},
				}
			}
		}
	}

	return s.model, nil
}
