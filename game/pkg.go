package game

import (
	"encoding/json"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/player"
	"github.com/ascii-arcade/farkle/score"
	"github.com/charmbracelet/lipgloss"
)

type Game struct {
	Players    []*player.Player `json:"players"`
	Scores     map[string]int   `json:"scores"`
	Turn       int              `json:"turn"`
	Round      int              `json:"round"`
	DicePool   dice.DicePool    `json:"dice_pool"`
	DiceHeld   dice.DicePool    `json:"dice_held"`
	DiceLocked []dice.DicePool  `json:"dice_locked"`
	LobbyCode  string           `json:"lobby_code"`
	Log        []string         `json:"log"`
	FirstRoll  bool             `json:"first_roll"`
	Busted     bool             `json:"busted"`

	Rolled bool
}

func New(lobbyCode string, players []*player.Player) *Game {
	scores := make(map[string]int)
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

	for i, p := range players {
		if p == nil {
			continue
		}
		scores[p.Id] = 0
		p.Color = colors[i]
	}

	return &Game{
		Players:   players,
		Scores:    scores,
		Turn:      0,
		Round:     1,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		LobbyCode: lobbyCode,
		FirstRoll: true,
	}
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
	if g.Turn >= len(g.Players) || g.Players[g.Turn] == nil {
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
	g.Log = append(g.Log, g.Players[g.Turn].StyledPlayerName()+" rolled: "+g.DicePool.RenderCharacters())

	if g.busted() {
		g.Busted = true
		g.Log = append(g.Log, g.Players[g.Turn].StyledPlayerName()+" busted!")
		g.NextTurn()
	}
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
	g.Log = append(g.Log, g.Players[g.Turn].StyledPlayerName()+" undid: "+dice.GetDieCharacter(lastDie))
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
	g.Log = append(g.Log, g.Players[g.Turn].StyledPlayerName()+" locked: "+g.DiceLocked[len(g.DiceLocked)-1].RenderCharacters())
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

	if g.Scores[g.Players[g.Turn].Id] == 0 && turnScore < 500 {
		return false
	}

	g.Scores[g.Players[g.Turn].Id] += turnScore
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.NextTurn()
	g.Log = append(g.Log, g.Players[g.Turn].StyledPlayerName()+" banked: "+strconv.Itoa(turnScore))
	return true
}

func (g *Game) IsTurn(p *player.Player) bool {
	return g.Players[g.Turn].Id == p.Id
}

func (g *Game) ToJSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (g *Game) PlayerScores() string {
	scores := make([]string, len(g.Players))

	for i, player := range g.Players {
		if player == nil {
			continue
		}

		content := g.StyledPlayerName(i) + ": " + strconv.Itoa(g.Scores[player.Id])
		isCurrentPlayer := g.Turn == i

		scores[i] = lipgloss.NewStyle().
			PaddingRight(2).
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(content)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		scores...,
	)
}

func (g *Game) StyledPlayerName(i int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(g.Players[i].Color))

	return style.Render(g.Players[i].Name)
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
