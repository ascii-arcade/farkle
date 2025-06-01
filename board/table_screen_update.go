package board

import (
	"slices"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/score"
	tea "github.com/charmbracelet/bubbletea"
)

func (s *tableScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	if !s.model.game.IsTurn(s.model.player) {
		return s.model, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if s.model.game.IsGameOver() && s.model.player.Host {
				s.model.game.Restart()
				return s.model, nil
			}

			if !s.model.game.Rolled && !s.model.rolling {
				s.model.rollTickCount = 0
				s.model.rolling = true
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
		case "s":
			if !s.model.game.Started && s.model.player.Host {
				s.model.game.Start()
			}
		case "l":
			_, err := s.model.game.DiceHeld.Score()
			if len(s.model.game.DiceHeld) != 0 && err == nil {
				s.model.game.LockDice()
			}
		case "y", "b":
			if len(s.model.game.DiceHeld) == 0 && len(s.model.game.DiceLocked) > 0 {
				s.model.game.Bank()
			}
		case "a":
			for _, face := range score.GetScorableDieFaces(s.model.game.DicePool) {
				s.model.game.HoldDie(face)
			}
		case "u", "backspace":
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.Undo()
			}
		case "c":
			s.model.game.ClearHeld()
		}
	}

	return s.model, nil
}
