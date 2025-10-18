package games

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/score"
)

type ActionType string

const (
	ActionTypeStart  ActionType = "start"
	ActionTypeRoll   ActionType = "roll"
	ActionTypeHold   ActionType = "hold"
	ActionTypeUndo   ActionType = "undo"
	ActionTypeLock   ActionType = "lock"
	ActionTypeBank   ActionType = "bank"
	ActionTypeBusted ActionType = "busted"
)

type Action struct {
	GameID     string          `bson:"game_id"`
	PlayerID   string          `bson:"player_id"`
	Action     ActionType      `bson:"action"`
	DicePool   dice.DicePool   `bson:"dice_pool,omitempty"`
	DiceLocked []dice.DicePool `bson:"dice_locked,omitempty"`
	Turn       int             `bson:"turn"`
	Timestamp  time.Time       `bson:"time"`
}

func (a *Action) Save() error {
	collection := database.GetDB().Collection(database.CollectionActions)
	_, err := collection.InsertOne(database.GetDB().Context(), a)
	return err
}

func (g *Game) saveAction(actionType ActionType) {
	action := Action{
		GameID:     g.Id,
		PlayerID:   g.GetTurnPlayer().Id,
		Action:     actionType,
		DicePool:   g.dicePool,
		DiceLocked: g.diceLocked,
		Turn:       g.Turn,
		Timestamp:  time.Now().UTC(),
	}
	if err := action.Save(); err != nil {
		slog.Error("error saving action", "error", err)
	}

	if err := g.Save(); err != nil {
		slog.Error("error saving game after action", "error", err)
	}
}

func (g *Game) Start() error {
	return g.withErrLock(func() error {
		if g.InProgress {
			return ErrGameAlreadyInProgress
		}

		g.randomizeTurnOrder()
		g.InProgress = true
		g.saveAction(ActionTypeStart)
		return nil
	})
}

func (g *Game) Restart() {
	g.withLock(func() {
		g.dicePool = dice.NewDicePool(6)
		g.diceHeld = dice.NewDicePool(0)
		g.diceLocked = []dice.DicePool{}
		g.firstRoll = true
		g.rolled = false
		g.endGame = false
		g.Turn = 0
		g.log = []string{}

		for _, p := range g.players {
			p.Score = 0
			p.PlayedLastTurn = false
		}
	})
}

func (g *Game) RollDice(rolling bool) {
	g.withLock(func() {
		g.dicePool.Roll()
		if !rolling {
			g.rolled = true
			g.log = append(g.log, g.players[g.GetTurnPlayer()].StyledPlayerName(g.style)+" rolled: "+g.dicePool.RenderCharacters())
			g.diceLocked = append(g.diceLocked, dice.NewDicePool(0))

			if g.Busted() {
				lockedScore := 0
				for _, diePool := range g.diceLocked {
					ls, _, _ := diePool.Score()
					lockedScore += ls
				}
				text := fmt.Sprintf("%s busted! (%d)", g.players[g.GetTurnPlayer()].StyledPlayerName(g.style), lockedScore)
				g.log = append(g.log, text)
				g.nextTurn()
				g.saveAction(ActionTypeBusted)
				return
			}

			g.saveAction(ActionTypeRoll)
		}
	})
}

func (g *Game) HoldDie(dieToHold int) {
	g.withLock(func() {
		if g.dicePool.Remove(dieToHold) {
			g.diceHeld.Add(dieToHold)
		}
		g.saveAction(ActionTypeHold)
	})
}

func (g *Game) Undo() {
	g.withLock(func() {
		if len(g.diceHeld) == 0 {
			return
		}
		lastDie := g.diceHeld[len(g.diceHeld)-1]
		if g.diceHeld.Remove(lastDie) {
			g.dicePool.Add(lastDie)
		}
		g.saveAction(ActionTypeUndo)
	})
}

func (g *Game) UndoAll() {
	for range len(g.diceHeld) {
		g.Undo()
	}
}

func (g *Game) LockDice() error {
	return g.withErrLock(func() error {
		if len(g.diceHeld) == 0 {
			return nil
		}
		if _, _, err := score.Calculate(g.diceHeld, false); err != nil {
			return err
		}

		if len(g.diceLocked) == 0 {
			g.diceLocked = append(g.diceLocked, dice.NewDicePool(0))
		}

		g.diceLocked[len(g.diceLocked)-1] = g.diceHeld.Copy()
		g.diceHeld = dice.NewDicePool(0)
		g.rolled = false

		if len(g.dicePool) == 0 {
			g.dicePool = dice.NewDicePool(6)
		}
		g.log = append(g.log, g.players[g.GetTurnPlayer()].StyledPlayerName(g.style)+" locked: "+g.diceLocked[len(g.diceLocked)-1].RenderCharacters())
		g.saveAction(ActionTypeLock)
		return nil
	})
}

func (g *Game) Bank() error {
	return g.withErrLock(func() error {
		turnScore := 0
		for _, diceLocked := range g.diceLocked {
			score, _, err := diceLocked.Score()
			if err != nil {
				return err
			}
			turnScore += score
		}

		p := g.GetTurnPlayer()
		if g.players[p].Score == 0 && turnScore < 500 {
			return ErrScoreTooLow
		}
		g.players[p].Score += turnScore

		g.diceHeld = dice.NewDicePool(0)
		g.diceLocked = []dice.DicePool{}
		g.log = append(g.log, g.players[p].StyledPlayerName(g.style)+" banked: "+strconv.Itoa(turnScore))

		g.nextTurn()

		g.saveAction(ActionTypeBank)
		return nil
	})
}
