package games

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"slices"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/score"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/xid"
)

type Game struct {
	DicePool   dice.DicePool
	DiceHeld   dice.DicePool
	DiceLocked []dice.DicePool
	Busted     bool
	Code       string
	Started    bool

	round     int
	firstRoll bool
	rolled    bool
	turn      int
	log       []string
	players   []*Player
	style     lipgloss.Style
	mu        sync.Mutex
}

var games = make(map[string]*Game)

func New(style lipgloss.Style) *Game {
	game := &Game{
		turn:      0,
		round:     1,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		firstRoll: true,
		Code:      utils.GenerateCode(),
		style:     style,
	}
	games[game.Code] = game
	return game
}

func (g *Game) Lock() {
	g.mu.Lock()
}
func (g *Game) Unlock() {
	g.mu.Unlock()
}

func Exists(code string) bool {
	_, ok := games[code]
	return ok
}

func Get(code string) (*Game, bool) {
	game, ok := games[code]
	return game, ok
}

func (g *Game) Start() {
	g.Started = true
}

func (g *Game) AddPlayer(host bool) *Player {
	player := &Player{
		Id:         xid.New().String(),
		Name:       utils.GenerateName(),
		UpdateChan: make(chan any, 1),
		Host:       host,
	}

	g.Lock()
	defer g.Unlock()

	g.players = append(g.players, player)
	g.Refresh()
	return player
}

func (g *Game) RemovePlayer(player *Player) {
	g.Lock()
	defer g.Unlock()

	for i, p := range g.players {
		if p.Id == player.Id {
			g.players = slices.Delete(g.players, i, i+1)
			break
		}
	}
	if len(g.players) == 0 {
		delete(games, g.Code)
		return
	}
}

func (g *Game) GetPlayers() []*Player {
	return g.players
}

func (g *Game) NextTurn() {
	g.turn++
	if g.turn >= len(g.players) {
		g.turn = 0
		g.round++
	}
	g.firstRoll = true
	g.rolled = false
	g.Busted = false
}

func (g *Game) GetHost() *Player {
	for _, p := range g.players {
		if p.Host {
			return p
		}
	}

	return nil
}

func (g *Game) RollDice() {
	g.Lock()
	defer g.Unlock()

	if g.firstRoll {
		g.DicePool = dice.NewDicePool(6)
		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = make([]dice.DicePool, 0)
		g.firstRoll = false
	}

	g.DicePool.Roll()
	g.rolled = true
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" rolled: "+g.DicePool.RenderCharacters())

	if g.busted() {
		g.Busted = true
		g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" busted!")
		g.NextTurn()
	}
}

func (g *Game) getTurnPlayer() *Player {
	g.Lock()
	defer g.Unlock()

	for _, p := range g.players {
		if p.TurnOrder == g.turn {
			return p
		}
	}
	return nil
}

func (g *Game) HoldDie(dieToHold int) {
	g.Lock()
	defer g.Unlock()

	if g.DicePool.Remove(dieToHold) {
		g.DiceHeld.Add(dieToHold)
	}
}

func (g *Game) Undo() {
	g.Lock()
	defer g.Unlock()

	if len(g.DiceHeld) == 0 {
		return
	}
	lastDie := g.DiceHeld[len(g.DiceHeld)-1]
	if g.DiceHeld.Remove(lastDie) {
		g.DicePool.Add(lastDie)
	}
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" undid: "+dice.GetDieCharacter(lastDie))
}

func (g *Game) LockDice() {
	g.Lock()
	defer g.Unlock()

	if len(g.DiceHeld) == 0 {
		return
	}

	_, scoreable := score.Calculate(g.DiceHeld)
	if !scoreable {
		return
	}

	g.DiceLocked = append(g.DiceLocked, g.DiceHeld)
	g.DiceHeld = dice.NewDicePool(0)
	g.rolled = false

	if len(g.DicePool) == 0 {
		g.DicePool = dice.NewDicePool(6)
	}
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())
}

func (g *Game) Bank() bool {
	g.Lock()
	defer g.Unlock()

	turnScore := 0
	for _, diceLocked := range g.DiceLocked {
		score, ok := diceLocked.Score()
		if !ok {
			return false
		}
		turnScore += score
	}

	p := g.getTurnPlayer()
	if p.Score == 0 && turnScore < 500 {
		return false
	}

	p.Score += turnScore
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.NextTurn()
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" banked: "+strconv.Itoa(turnScore))
	return true
}

func (g *Game) IsTurn(p *Player) bool {
	return g.getTurnPlayer().Id == p.Id
}

func (g *Game) ToJSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (g *Game) PlayerScores() string {
	scores := make([]string, 0, len(g.players))

	for _, p := range g.players {
		if p == nil {
			continue
		}

		isCurrentPlayer := g.turn == p.TurnOrder
		scores = append(scores, g.style.
			PaddingRight(2).
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(g.getTurnPlayer().StyledPlayerName(g.style)+": "+strconv.Itoa(p.Score)))
	}

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
	_, ok := score.Calculate(g.DicePool)
	return !ok
}

func (g *Game) Refresh() {
	for _, p := range g.players {
		select {
		case p.UpdateChan <- struct{}{}:
		default:
		}
	}
}

func (g *Game) GetPlayer(id string) *Player {
	g.Lock()
	defer g.Unlock()

	for _, p := range g.players {
		if p.Id == id {
			return p
		}
	}

	return nil
}
