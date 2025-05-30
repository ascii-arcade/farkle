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
	FirstRoll  bool
	Rolled     bool

	round   int
	turn    int
	log     []string
	players []*Player
	style   lipgloss.Style
	mu      sync.Mutex
}

var games = make(map[string]*Game)

func New(style lipgloss.Style) *Game {
	game := &Game{
		turn:      0,
		round:     1,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		FirstRoll: true,
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
	g.Lock()
	defer g.Unlock()
	if g.Started {
		return
	}
	g.Started = true
	g.Refresh()
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
	g.FirstRoll = true
	g.Rolled = false
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

	if g.FirstRoll {
		g.DicePool = dice.NewDicePool(6)
		g.DiceHeld = dice.NewDicePool(0)
		g.DiceLocked = make([]dice.DicePool, 0)
		g.FirstRoll = false
	}

	g.DicePool.Roll()
	g.Rolled = true
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" rolled: "+g.DicePool.RenderCharacters())

	if g.busted() {
		g.Busted = true
		g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" busted!")
		g.NextTurn()
	}

	g.Refresh()
}

func (g *Game) getTurnPlayer() *Player {
	for i, p := range g.players {
		if i == g.turn {
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

	g.Refresh()
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

	g.Refresh()
}

func (g *Game) LockDice() {
	g.Lock()
	defer g.Unlock()

	if len(g.DiceHeld) == 0 {
		return
	}
	if _, err := score.Calculate(g.DiceHeld, false); err != nil {
		return
	}

	g.DiceLocked = append(g.DiceLocked, g.DiceHeld)
	g.DiceHeld = dice.NewDicePool(0)
	g.Rolled = false

	if len(g.DicePool) == 0 {
		g.DicePool = dice.NewDicePool(6)
	}
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())

	g.Refresh()
}

func (g *Game) Bank() {
	g.Lock()
	defer g.Unlock()

	turnScore := 0
	for _, diceLocked := range g.DiceLocked {
		score, err := diceLocked.Score()
		if err != nil {
			return
		}
		turnScore += score
	}

	p := g.getTurnPlayer()
	if p.Score == 0 && turnScore < 500 {
		return
	}
	p.Score += turnScore

	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.NextTurn()
	g.log = append(g.log, g.getTurnPlayer().StyledPlayerName(g.style)+" banked: "+strconv.Itoa(turnScore))

	g.Refresh()
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

	for i, p := range g.players {
		if p == nil {
			continue
		}

		isCurrentPlayer := g.turn == i
		scores = append(scores, g.style.
			PaddingRight(2).
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(p.StyledPlayerName(g.style)+": "+strconv.Itoa(p.Score)))
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
