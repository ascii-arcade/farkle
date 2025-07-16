package board

import (
	"slices"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/score"
	tea "github.com/charmbracelet/bubbletea"
)

func (s *tableScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case rollMsg:
		if s.rollTickCount < rollFrames {
			s.rollTickCount++
			s.model.game.RollDice(s.rolling)
			return s.model, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		s.rolling = false
		s.model.game.RollDice(s.rolling)

	case tea.KeyMsg:
		if keys.OpenHelp.TriggeredBy(msg.String()) {
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: &helpScreen{model: s.model},
				}
			}
		}

		if !s.model.game.IsTurn(s.model.player) {
			return s.model, nil
		}
		s.model.error = ""

		playerData := s.model.game.GetPlayerData(s.model.player)

		if keys.ActionRoll.TriggeredBy(msg.String()) {
			if s.model.game.IsGameOver() && playerData.IsHost {
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
		}

		if keys.LobbyStartGame.TriggeredBy(msg.String()) {
			if s.model.game.Ready() && playerData.IsHost {
				if err := s.model.game.Start(); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
				}
			}
		}

		if keys.ActionLock.TriggeredBy(msg.String()) {
			_, _, err := s.model.game.DiceHeld.Score()
			if len(s.model.game.DiceHeld) != 0 && err == nil {
				if err := s.model.game.LockDice(); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}
			}
		}

		if keys.ActionBank.TriggeredBy(msg.String()) {
			if s.model.game.Rolled && len(s.model.game.DiceLocked[len(s.model.game.DiceLocked)-1]) == 0 {
				s.model.error = s.model.lang().Get("error", "game", "lock_before_banking")
				return s.model, nil
			}
			if len(s.model.game.DiceHeld) == 0 && len(s.model.game.DiceLocked) > 0 {
				if err := s.model.game.Bank(); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}
			}
		}

		if keys.ActionTakeAll.TriggeredBy(msg.String()) {
			if !s.model.game.Rolled {
				s.model.error = s.model.lang().Get("error", "game", "did_not_roll")
				return s.model, nil
			}
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.UndoAll()
			}
			_, all, _ := score.Calculate(s.model.game.DicePool, true)
			allCopy := slices.Clone(all)
			for _, face := range allCopy {
				s.model.game.HoldDie(face)
			}
		}

		if keys.ActionUndo.TriggeredBy(msg.String()) {
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.Undo()
			}
		}

		if keys.ActionUndoAll.TriggeredBy(msg.String()) {
			if len(s.model.game.DiceHeld) > 0 {
				s.model.game.UndoAll()
			}
		}

		switch msg.String() {
		case "1", "2", "3", "4", "5", "6":
			if !s.model.game.Rolled {
				s.model.error = s.model.lang().Get("error", "game", "did_not_roll")
				return s.model, nil
			}

			if slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, msg.String()) {
				face, _ := strconv.Atoi(msg.String())
				if s.model.game.DicePool.Contains(face) {
					s.model.game.HoldDie(face)
					return s.model, nil
				}
			}

		case "!", "@", "#", "$", "%", "^":
			if !s.model.game.Rolled {
				s.model.error = s.model.lang().Get("error", "game", "did_not_roll")
				return s.model, nil
			}

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
	}

	return s.model, nil
}
