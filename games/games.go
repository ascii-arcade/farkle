package games

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
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

	endGame bool
	colors  []string
	turn    int
	log     []string
	players []*Player
	style   lipgloss.Style
	mu      sync.Mutex
}

var games = make(map[string]*Game)

func New(style lipgloss.Style) *Game {
	colors := []string{
		"#3B82F6", // Blue
		"#10B981", // Green
		"#FACC15", // Yellow
		"#8B5CF6", // Purple
		"#06B6D4", // Cyan
		"#F97316", // Orange
	}

	rand.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	game := &Game{
		turn:      0,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		FirstRoll: true,
		Code:      utils.GenerateCode(),
		style:     style,
		colors:    colors,
	}
	games[game.Code] = game
	return game
}

func Exists(code string) bool {
	_, ok := games[code]
	return ok
}

func Get(code string) (*Game, bool) {
	game, ok := games[code]
	return game, ok
}

func GetAll() []*Game {
	gamesList := make([]*Game, 0, len(games))
	for _, game := range games {
		gamesList = append(gamesList, game)
	}
	return gamesList
}

func (g *Game) Lock() {
	g.mu.Lock()
}
func (g *Game) Unlock() {
	g.mu.Unlock()
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
		UpdateChan: make(chan struct{}, 1),
		Host:       host,
		Color:      g.colors[len(g.players)%len(g.colors)],
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

	if len(g.players) > 0 && player.Host {
		g.players[0].Host = true
	}

	if len(g.players) == 1 && g.Started {
		g.Started = false
	}

	if len(g.players) == 0 {
		delete(games, g.Code)
		return
	}

	defer g.Refresh()
}

func (g *Game) GetPlayers() []*Player {
	g.Lock()
	defer g.Unlock()
	return g.players
}

func (g *Game) Restart() {
	g.Lock()
	defer g.Unlock()

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

	g.Refresh()
}

func (g *Game) NextTurn() {
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
		return
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

	g.DicePool.Roll()
	g.Rolled = true
	g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" rolled: "+g.DicePool.RenderCharacters())

	if g.busted() {
		g.Busted = true
		g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" busted!")
		g.NextTurn()
	}

	g.Refresh()
}

func (g *Game) GetTurnPlayer() *Player {
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

func (g *Game) ClearHeld() {
	g.Lock()
	defer g.Unlock()

	if len(g.DiceHeld) == 0 {
		return
	}

	for _, die := range g.DiceHeld {
		g.DicePool.Add(die)
	}
	g.DiceHeld = dice.NewDicePool(0)
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

func (g *Game) UndoAll() {
	for range len(g.DiceHeld) {
		g.Undo()
	}
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
	g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())

	g.Refresh()
}

func (g *Game) Bank() error {
	g.Lock()
	defer g.Unlock()

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
		return errors.New("need to bank 500 or more before you can bank less")
	}
	p.Score += turnScore

	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.log = append(g.log, g.GetTurnPlayer().StyledPlayerName(g.style)+" banked: "+strconv.Itoa(turnScore))

	g.NextTurn()
	g.Refresh()

	return nil
}

func (g *Game) GetWinningPlayer() *Player {
	if !g.endGame {
		return nil
	}

	var winningPlayer *Player
	for _, p := range g.players {
		if winningPlayer == nil || p.Score > winningPlayer.Score {
			winningPlayer = p
		}
	}

	return winningPlayer
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
	return g.GetTurnPlayer().Id == p.Id
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
		isWinner := g.GetWinningPlayer() != nil && g.GetWinningPlayer().Id == p.Id
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
