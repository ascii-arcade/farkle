package game

import (
	"encoding/json"
	"math/rand/v2"
	"strconv"

	"github.com/ascii-arcade/farkle/internal/dice"
	"github.com/ascii-arcade/farkle/internal/player"
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
	Rolling    bool             `json:"rolling"`
	Log        []string         `json:"log"`

	roll chan struct{}

	log log
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
	}
}

func (g *Game) Roll() {
	g.Rolling = true
	g.roll <- struct{}{}
}

func (g *Game) Update(gIn Game) {
	g.Players = gIn.Players
	g.Scores = gIn.Scores
	g.Turn = gIn.Turn
	g.Round = gIn.Round
	g.DicePool = gIn.DicePool
	g.DiceHeld = gIn.DiceHeld
	g.DiceLocked = gIn.DiceLocked
	g.Rolling = gIn.Rolling
}

func (g *Game) NextTurn() {
	g.Turn++
	if g.Turn >= len(g.Players) {
		g.Turn = 0
		g.Round++
	}
}

func (g *Game) RollDice() {
	g.Rolling = true
	g.DicePool.Roll()
}

func (g *Game) HoldDie(dieToHold int) {
	g.DiceHeld.Add(dieToHold)
	g.DicePool.Remove(dieToHold)
}

func (g *Game) Undo() {
	lastDie := g.DiceHeld[len(g.DiceHeld)-1]
	g.DiceHeld.Remove(lastDie)
	g.DicePool.Add(lastDie)
}

func (g *Game) LockDice() {
	if len(g.DiceHeld) == 0 {
		return
	}

	g.DiceLocked = append(g.DiceLocked, g.DiceHeld)
	g.DiceHeld = dice.NewDicePool(0)
}

func (g *Game) Bank() error {
	for _, diceLocked := range g.DiceLocked {
		score, err := diceLocked.Score()
		if err != nil {
			return err
		}
		g.Scores[g.Players[g.Turn].Id] += score
	}
	g.DicePool = dice.NewDicePool(6)
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}

	return nil
}

func (g *Game) Bust() {
	g.DicePool = dice.NewDicePool(6)
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = []dice.DicePool{}
	g.NextTurn()
}

func (g *Game) IsTurn(p *player.Player) bool {
	return g.Players[g.Turn].Id == p.Id
}

func (g *Game) ToJSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (g *Game) playerScores() string {
	scores := make([]string, len(g.Players))

	for i, player := range g.Players {
		if player == nil {
			continue
		}

		content := g.styledPlayerName(i) + ": " + strconv.Itoa(g.Scores[player.Id])
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

func (g *Game) styledPlayerName(i int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(g.Players[i].Color))

	return style.Render(g.Players[i].Name)
}
