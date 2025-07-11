package games

import (
	"strconv"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/score"
)

func (g *Game) Start() error {
	return g.withErrLock(func() error {
		if g.InProgress {
			return ErrGameAlreadyInProgress
		}
		for _, p := range g.players {
			p.Score = 0
			p.PlayedLastTurn = false
		}
		g.InProgress = true
		return nil
	})
}

func (g *Game) Restart() {
	g.withLock(func() {
		g.DicePool = dice.NewDicePool(6)
		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = []dice.DicePool{}
		g.Busted = false
		g.FirstRoll = true
		g.Rolled = false
		g.endGame = false
		g.turn = 0
		g.log = []string{}

		for _, p := range g.players {
			p.Score = 0
			p.PlayedLastTurn = false
		}
	})
}

func (g *Game) RollDice(rolling bool) {
	g.withLock(func() {
		g.DicePool.Roll()
		if !rolling {
			g.Rolled = true
			g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" rolled: "+g.DicePool.RenderCharacters())

			if g.busted() {
				g.Busted = true
				g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" busted!")
				g.nextTurn()
			}
		}
	})
}

func (g *Game) HoldDie(dieToHold int) {
	g.withLock(func() {
		if g.DicePool.Remove(dieToHold) {
			g.DiceHeld.Add(dieToHold)
		}
	})
}

func (g *Game) Undo() {
	g.withLock(func() {
		if len(g.DiceHeld) == 0 {
			return
		}
		lastDie := g.DiceHeld[len(g.DiceHeld)-1]
		if g.DiceHeld.Remove(lastDie) {
			g.DicePool.Add(lastDie)
		}
	})
}

func (g *Game) UndoAll() {
	for range len(g.DiceHeld) {
		g.Undo()
	}
}

func (g *Game) LockDice() error {
	return g.withErrLock(func() error {
		if len(g.DiceHeld) == 0 {
			return nil
		}
		if _, _, err := score.Calculate(g.DiceHeld, false); err != nil {
			return err
		}

		g.DiceLocked = append(g.DiceLocked, g.DiceHeld)
		g.DiceHeld = dice.NewDicePool(0)
		g.Rolled = false

		if len(g.DicePool) == 0 {
			g.DicePool = dice.NewDicePool(6)
		}
		g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())
		return nil
	})
}

func (g *Game) Bank() error {
	return g.withErrLock(func() error {
		turnScore := 0
		for _, diceLocked := range g.DiceLocked {
			score, _, err := diceLocked.Score()
			if err != nil {
				return err
			}
			turnScore += score
		}

		p := g.GetTurnPlayer()
		if p.Score == 0 && turnScore < 500 {
			return ErrScoreTooLow
		}
		p.Score += turnScore

		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = []dice.DicePool{}
		g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" banked: "+strconv.Itoa(turnScore))

		g.nextTurn()
		return nil
	})
}
