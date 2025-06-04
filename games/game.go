package games

import (
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/score"
	"github.com/charmbracelet/lipgloss"
)

type Game struct {
	DicePool   dice.DicePool
	DiceHeld   dice.DicePool
	DiceLocked []dice.DicePool
	Busted     bool
	Code       string
	InProgress bool
	FirstRoll  bool
	Rolled     bool

	endGame bool
	colors  []string
	turn    int
	log     []string
	players []*Player
	style   lipgloss.Style
	mu      sync.Mutex
}

func (g *Game) withLock(fn func() error) error {
	g.mu.Lock()
	defer func() {
		g.Refresh()
		g.mu.Unlock()
	}()
	return fn()
}

func (g *Game) Start() error {
	return g.withLock(func() error {
		if g.InProgress {
			return ErrGameAlreadyInProgress
		}
		g.InProgress = true
		return nil
	})
}

func (g *Game) AddPlayer(player *Player, isHost bool) error {
	return g.withLock(func() error {
		if g.InProgress {
			return ErrGameAlreadyInProgress
		}

		go func() {
			<-player.ctx.Done()
			g.RemovePlayer(player)
		}()

		if isHost {
			player.IsHost = true
		}

		g.players = append(g.players, player)
		return nil
	})
}

func (g *Game) RemovePlayer(player *Player) {
	_ = g.withLock(func() error {
		for i, p := range g.players {
			if p.Name == player.Name {
				g.players = slices.Delete(g.players, i, i+1)
				break
			}
		}

		if len(g.players) > 0 && player.IsHost {
			g.players[0].IsHost = true
		}

		if len(g.players) == 1 && g.InProgress {
			g.InProgress = false
		}

		if len(g.players) == 0 {
			delete(games, g.Code)
		}

		return nil
	})
}

func (g *Game) GetPlayers() []*Player {
	return g.players
}

func (g *Game) Restart() {
	_ = g.withLock(func() error {
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
		return nil
	})
}

func (g *Game) Ready() bool {
	return len(g.players) >= 2 && !g.InProgress
}

func (g *Game) NextTurn() {
	_ = g.withLock(func() error {
		player := g.GetTurnPlayer()
		if player.Score >= 10000 && !g.endGame {
			g.endGame = true
			g.log = append(g.log, player.StyledPlayerName(g.style)+" triggered end game!")
		}

		if g.endGame && !player.PlayedLastTurn {
			player.PlayedLastTurn = true
		}

		if g.IsGameOver() {
			winner := g.GetWinningPlayer()
			g.log = append(g.log, winner.StyledPlayerName(g.style)+" wins the game with a score of "+strconv.Itoa(winner.Score)+"!")
			return nil
		}

		g.turn++
		if g.turn >= len(g.players) {
			g.turn = 0
		}
		g.FirstRoll = true
		g.Rolled = false
		g.Busted = false
		g.DiceLocked = []dice.DicePool{}
		g.DicePool = dice.NewDicePool(6)
		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = make([]dice.DicePool, 0)
		g.FirstRoll = false

		return nil
	})
}

func (g *Game) GetHost() *Player {
	var player *Player
	_ = g.withLock(func() error {
		for _, p := range g.players {
			if p.IsHost {
				player = p
				break
			}
		}

		return nil
	})
	return player
}

func (g *Game) RollDice() {
	_ = g.withLock(func() error {
		g.DicePool.Roll()
		g.Rolled = true
		g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" rolled: "+g.DicePool.RenderCharacters())

		if g.busted() {
			g.Busted = true
			g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" busted!")
			g.NextTurn()
		}
		return nil
	})
}

func (g *Game) GetTurnPlayer() *Player {
	var player *Player
	_ = g.withLock(func() error {
		for i, p := range g.players {
			if i == g.turn {
				player = p
				break
			}
		}
		return nil
	})
	return player
}

func (g *Game) HoldDie(dieToHold int) {
	_ = g.withLock(func() error {
		if g.DicePool.Remove(dieToHold) {
			g.DiceHeld.Add(dieToHold)
		}
		return nil
	})
}

func (g *Game) ClearHeld() error {
	return g.withLock(func() error {
		if len(g.DiceHeld) == 0 {
			return ErrNoDiceHeld
		}

		for _, die := range g.DiceHeld {
			g.DicePool.Add(die)
		}
		g.DiceHeld = dice.NewDicePool(0)

		return nil
	})
}

func (g *Game) Undo() {
	g.withLock(func() error {
		if len(g.DiceHeld) == 0 {
			return nil
		}
		lastDie := g.DiceHeld[len(g.DiceHeld)-1]
		if g.DiceHeld.Remove(lastDie) {
			g.DicePool.Add(lastDie)
		}
		return nil
	})
}

func (g *Game) UndoAll() {
	for range len(g.DiceHeld) {
		g.Undo()
	}
}

func (g *Game) LockDice() error {
	return g.withLock(func() error {
		if len(g.DiceHeld) == 0 {
			return nil
		}
		if _, err := score.Calculate(g.DiceHeld, false); err != nil {
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
	return g.withLock(func() error {

		turnScore := 0
		for _, diceLocked := range g.DiceLocked {
			score, err := diceLocked.Score()
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

		g.NextTurn()
		return nil
	})
}

func (g *Game) GetWinningPlayer() *Player {
	var player *Player
	_ = g.withLock(func() error {
		if !g.endGame {
			return nil
		}

		for _, p := range g.players {
			if player == nil || p.Score > player.Score {
				player = p
			}
		}

		return nil
	})

	return player
}

func (g *Game) IsGameOver() bool {
	if !g.endGame {
		return false
	}

	for _, p := range g.players {
		if !p.PlayedLastTurn {
			return false
		}
	}

	return true
}

func (g *Game) IsTurn(p *Player) bool {
	return g.GetTurnPlayer().Name == p.Name
}

func (g *Game) PlayerScores() string {
	scores := make([]string, 0, len(g.players))

	_ = g.withLock(func() error {
		for i, p := range g.players {
			if p == nil {
				continue
			}

			isCurrentPlayer := g.turn == i
			isWinner := g.GetWinningPlayer() != nil && g.GetWinningPlayer().Name == p.Name
			playerName := p.StyledPlayerName(g.style)
			if isWinner {
				playerName = "★" + playerName + "★"
			}
			scores = append(scores, g.style.
				PaddingRight(2).
				Bold(isCurrentPlayer).
				Italic(isCurrentPlayer).
				Render(playerName+": "+strconv.Itoa(p.Score)))
		}
		return nil
	})

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		scores...,
	)
}

func (g *Game) RenderLog(limit int) string {
	if limit <= 0 || len(g.log) == 0 {
		return ""
	}
	if limit >= len(g.log) {
		return strings.Join(g.log, "\n")
	}
	return strings.Join(g.log[len(g.log)-limit:], "\n")
}

func (g *Game) busted() bool {
	if _, err := score.Calculate(g.DicePool, true); err != nil {
		return true
	}
	return false
}

func (g *Game) Refresh() {
	for _, p := range g.players {
		select {
		case p.UpdateChan <- struct{}{}:
		default:
		}
	}
}

func (g *Game) GetPlayer(name string) *Player {
	var player *Player
	_ = g.withLock(func() error {
		for _, p := range g.players {
			if p.Name == name {
				player = p
				break
			}
		}
		return nil
	})
	return player
}
