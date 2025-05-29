package game

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/player"
	"github.com/ascii-arcade/farkle/score"
	"github.com/charmbracelet/lipgloss"
)

type Game struct {
	Players map[string]*player.Player
	Clients map[chan any]struct{}

	mu sync.Mutex

	Scores     map[string]int
	Turn       int
	Round      int
	DicePool   dice.DicePool
	DiceHeld   dice.DicePool
	DiceLocked []dice.DicePool
	LobbyCode  string
	Log        []string
	FirstRoll  bool
	Busted     bool
	Rolled     bool
}

var Games = make(map[string]*Game)

func New() *Game {
	scores := make(map[string]int)

	return &Game{
		Scores:    scores,
		Turn:      0,
		Round:     1,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		FirstRoll: true,
	}
}

func (g *Game) Lock() {
	g.mu.Lock()
}
func (g *Game) Unlock() {
	g.mu.Unlock()
}

func (s *Game) AddClient(ch chan any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients[ch] = struct{}{}
}

func (s *Game) RemoveClient(ch chan any, player *player.Player) {
	delete(s.Clients, ch)
	delete(s.Players, player.Id)
}

func (g *Game) Update(gIn Game) {
	g.Players = gIn.Players
	g.Scores = gIn.Scores
	g.Turn = gIn.Turn
	g.Round = gIn.Round
	g.DicePool = gIn.DicePool
	g.DiceHeld = gIn.DiceHeld
	g.DiceLocked = gIn.DiceLocked
	g.Log = gIn.Log
	g.Rolled = gIn.Rolled
	g.FirstRoll = gIn.FirstRoll
}

func (g *Game) NextTurn() {
	g.Turn++
	if g.Turn >= len(g.Players) {
		g.Turn = 0
		g.Round++
	}
	g.FirstRoll = true
	g.Rolled = false
	g.Busted = false
}

func (g *Game) RollDice() {
	if g.FirstRoll {
		g.DicePool = dice.NewDicePool(6)
		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = make([]dice.DicePool, 0)
		g.FirstRoll = false
	}

	g.DicePool.Roll()
	g.Rolled = true
	g.Log = append(g.Log, g.getTurnPlayer().StyledPlayerName()+" rolled: "+g.DicePool.RenderCharacters())

	if g.busted() {
		g.Busted = true
		g.Log = append(g.Log, g.getTurnPlayer().StyledPlayerName()+" busted!")
		g.NextTurn()
	}
}

func (g *Game) getTurnPlayer() *player.Player {
	for _, p := range g.Players {
		if p.TurnOrder == g.Turn {
			return p
		}
	}
	return nil
}

func (g *Game) HoldDie(dieToHold int) {
	if g.DicePool.Remove(dieToHold) {
		g.DiceHeld.Add(dieToHold)
	}
}

func (g *Game) Undo() {
	if len(g.DiceHeld) == 0 {
		return
	}
	lastDie := g.DiceHeld[len(g.DiceHeld)-1]
	if g.DiceHeld.Remove(lastDie) {
		g.DicePool.Add(lastDie)
	}
	g.Log = append(g.Log, g.getTurnPlayer().StyledPlayerName()+" undid: "+dice.GetDieCharacter(lastDie))
}

func (g *Game) LockDice() {
	if len(g.DiceHeld) == 0 {
		return
	}

	_, scoreable := score.Calculate(g.DiceHeld)
	if !scoreable {
		return
	}

	g.DiceLocked = append(g.DiceLocked, g.DiceHeld)
	g.DiceHeld = dice.NewDicePool(0)
	g.Rolled = false

	if len(g.DicePool) == 0 {
		g.DicePool = dice.NewDicePool(6)
	}
	g.Log = append(g.Log, g.getTurnPlayer().StyledPlayerName()+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())
}

func (g *Game) Bank() bool {
	turnScore := 0
	for _, diceLocked := range g.DiceLocked {
		score, ok := diceLocked.Score()
		if !ok {
			return false
		}
		turnScore += score
	}

	if g.Scores[g.getTurnPlayer().Id] == 0 && turnScore < 500 {
		return false
	}

	g.Scores[g.getTurnPlayer().Id] += turnScore
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.NextTurn()
	g.Log = append(g.Log, g.getTurnPlayer().StyledPlayerName()+" banked: "+strconv.Itoa(turnScore))
	return true
}

func (g *Game) IsTurn(p *player.Player) bool {
	return g.getTurnPlayer().Id == p.Id
}

func (g *Game) ToJSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (g *Game) PlayerScores() string {
	scores := make([]string, 0, len(g.Players))

	for _, player := range g.Players {
		if player == nil {
			continue
		}
		isCurrentPlayer := g.Turn == player.TurnOrder

		scores = append(scores, lipgloss.NewStyle().
			PaddingRight(2).
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(g.getTurnPlayer().StyledPlayerName()+": "+strconv.Itoa(g.Scores[player.Id])))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		scores...,
	)
}

func (g *Game) RenderLog(limit int) string {
	if limit <= 0 || len(g.Log) == 0 {
		return ""
	}
	if limit >= len(g.Log) {
		return strings.Join(g.Log, "\n")
	}
	return strings.Join(g.Log[len(g.Log)-limit:], "\n")
}

func (g *Game) busted() bool {
	_, ok := score.Calculate(g.DicePool)
	return !ok
}

func (s *Game) Refresh() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.Clients {
		select {
		case ch <- 0:
		default:
		}
	}
}
