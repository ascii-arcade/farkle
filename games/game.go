package games

import (
	"log/slog"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/players"
	"github.com/ascii-arcade/farkle/score"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Game struct {
	Id         string     `bson:"_id"`
	Code       string     `bson:"code"`
	InProgress bool       `bson:"in_progress"`
	CreatedAt  *time.Time `bson:"created_at"`
	UpdatedAt  *time.Time `bson:"updated_at,omitempty"`
	EndedAt    *time.Time `bson:"ended_at,omitempty"`
	WinnerID   string     `bson:"winner_id,omitempty"`
	Turn       int        `bson:"turn"`

	dicePool   dice.DicePool
	diceHeld   dice.DicePool
	diceLocked []dice.DicePool
	firstRoll  bool
	rolled     bool
	endGame    bool
	status     GameStatus

	colors  []lipgloss.Color
	log     []string
	style   lipgloss.Style
	players map[*players.Player]*PlayerData

	mu sync.Mutex
}

func (g *Game) Save() error {
	g.UpdatedAt = utils.ToPointer(time.Now())
	collection := database.GetDB().Collection(database.CollectionGames)

	gameJson, err := g.toJson()
	if err != nil {
		return err
	}

	opts := options.Replace().SetUpsert(true)
	_, err = collection.ReplaceOne(database.GetDB().Context(), bson.M{"_id": g.Id}, gameJson, opts)
	return err
}

func (g *Game) toJson() (map[string]any, error) {
	playerIds := make([]string, 0, len(g.players))
	for p := range g.players {
		playerIds = append(playerIds, p.Id)
	}

	var gameMap map[string]any
	bytes, err := bson.Marshal(g)
	if err != nil {
		slog.Error("error marshalling game to json", "error", err)
		return nil, err
	}
	if err := bson.Unmarshal(bytes, &gameMap); err != nil {
		slog.Error("error unmarshalling game to map", "error", err)
		return nil, err
	}
	gameMap["player_ids"] = playerIds

	return gameMap, nil
}

func (g *Game) withErrLock(fn func() error) error {
	g.mu.Lock()
	defer func() {
		g.Refresh()
		g.mu.Unlock()
	}()
	return fn()
}

func (g *Game) withLock(fn func()) {
	g.mu.Lock()
	defer func() {
		g.Refresh()
		g.mu.Unlock()
	}()
	fn()
}

func (g *Game) Refresh() {
	for p := range g.players {
		select {
		case p.UpdateChan() <- struct{}{}:
		default:
		}
	}
}

func (g *Game) AddPlayer(player *players.Player) error {
	return g.withErrLock(func() error {
		if _, ok := g.getPlayer(player.Id); ok {
			return nil
		}

		if g.InProgress {
			return ErrGameAlreadyInProgress
		}

		playerData := &PlayerData{
			Name:      player.Username,
			Score:     0,
			Color:     g.colors[len(g.players)%len(g.colors)],
			turnOrder: len(g.players),
			InGame:    true,
		}

		if len(g.players) == 0 {
			playerData.IsHost = true
		}

		player.OnDisconnect(func() {
			if !g.InProgress {
				g.RemovePlayer(player)
			}
		})

		g.players[player] = playerData
		return g.Save()
	})
}

func (g *Game) RemovePlayer(player *players.Player) {
	g.withLock(func() {
		if len(g.players) > 0 && g.players[player].IsHost {
			for _, pd := range g.players {
				if !pd.IsHost {
					pd.IsHost = true
					break
				}
			}
		}

		if len(g.players) <= 1 && g.InProgress {
			g.InProgress = false
		}

		if g.GetPlayerCount(false) == 0 {
			g.EndedAt = utils.ToPointer(time.Now())
			if err := g.Save(); err != nil {
				slog.Error("error saving game after player removal", "error", err)
			}
			delete(games, g.Code)
			return
		}

		g.players[player].InGame = false
	})
}

func (g *Game) GetPlayers() []*players.Player {
	players := make([]*players.Player, 0, len(g.players))
	for p := range g.players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool {
		return g.players[players[i]].turnOrder < g.players[players[j]].turnOrder
	})
	return players
}

func (g *Game) GetHost() *players.Player {
	var player *players.Player
	g.withLock(func() {
		for p, pd := range g.players {
			if pd.IsHost {
				player = p
				break
			}
		}
	})
	return player
}

func (g *Game) getPlayer(id string) (*players.Player, bool) {
	for p := range g.players {
		if p.Id == id {
			return p, true
		}
	}
	return nil, false
}

func (g *Game) GetDisconnectedPlayers() []*players.Player {
	var players []*players.Player
	g.withLock(func() {
		for p := range g.players {
			if !p.IsConnected() {
				players = append(players, p)
			}
		}
	})
	return players
}

func (g *Game) HasPlayer(player *players.Player) bool {
	_, exists := g.getPlayer(player.Id)
	return exists
}

func (g *Game) GetPlayerCount(includeDisconnected bool) int {
	count := 0
	for p := range g.players {
		if includeDisconnected || p.IsConnected() {
			count++
		}
	}
	return count
}

func (g *Game) GetPlayerData(player *players.Player) *PlayerData {
	return g.players[player]
}

func (g *Game) randomizeTurnOrder() {
	nums := make([]int, len(g.players))
	for i := range nums {
		nums[i] = i
	}
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	i := 0
	for _, playerData := range g.players {
		playerData.turnOrder = nums[i]
		i++
	}
}

func (g *Game) GetStatus() GameStatus {
	return g.status
}

func (g *Game) Ready() bool {
	return len(g.players) >= 2 && !g.InProgress
}

func (g *Game) ValidGame() bool {
	return len(g.players) >= 2
}

func (g *Game) GetTurnPlayer() *players.Player {
	for p, pd := range g.players {
		if pd.turnOrder == g.Turn {
			return p
		}
	}
	return nil
}

func (g *Game) IsTurn(p *players.Player) bool {
	return g.players[g.GetTurnPlayer()].Name == g.players[p].Name
}

func (g *Game) nextTurn() {
	score := 10000
	if config.GetDebug() {
		score = 1000
	}

	player := g.GetTurnPlayer()
	playerData := g.players[player]
	if playerData.Score >= score && !g.endGame {
		g.endGame = true
		g.log = append(g.log, playerData.StyledPlayerName(g.style)+" triggered end game!")
	}

	if g.IsGameOver() {
		winner := g.GetWinningPlayer()
		winnerData := g.players[winner]
		g.log = append(g.log, winnerData.StyledPlayerName(g.style)+" wins the game with a score of "+strconv.Itoa(winnerData.Score)+"!")
		g.InProgress = false
		g.EndedAt = utils.ToPointer(time.Now())
		g.WinnerID = winner.Id
		if err := g.Save(); err != nil {
			slog.Error("error saving game after game over", "error", err)
		}
		return
	}

	if g.endGame && !playerData.PlayedLastTurn {
		playerData.PlayedLastTurn = true
	}

	g.Turn++
	if g.Turn >= len(g.players) {
		g.Turn = 0
	}
	g.rolled = false
	g.diceLocked = []dice.DicePool{}
	g.dicePool = dice.NewDicePool(6)
	g.diceHeld = dice.NewDicePool(0)
	g.diceLocked = make([]dice.DicePool, 0)
	g.firstRoll = false
}

func (g *Game) GetWinningPlayer() *players.Player {
	var winningPlayer *players.Player
	var winningPlayerData *PlayerData
	if !g.endGame {
		return nil
	}

	for _, p := range g.GetPlayers() {
		pd := g.GetPlayerData(p)
		if winningPlayer == nil {
			winningPlayer = p
			winningPlayerData = pd
			continue
		}

		if pd.Score > winningPlayerData.Score {
			winningPlayer = p
			winningPlayerData = pd
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

func (g *Game) PlayerScores() string {
	scores := make([]string, 0, len(g.players))

	for _, p := range g.GetPlayers() {
		pd := g.GetPlayerData(p)
		isCurrentPlayer := g.Turn == pd.turnOrder
		winningPlayer := g.GetWinningPlayer()
		isWinning := winningPlayer != nil && g.players[winningPlayer].Name == pd.Name
		playerName := pd.StyledPlayerName(g.style)
		if isWinning {
			playerName = "★" + playerName + "★"
		}
		scores = append(scores, g.style.
			PaddingRight(2).
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(playerName+": "+strconv.Itoa(pd.Score)))
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

func (g *Game) Busted() bool {
	if _, _, err := score.Calculate(g.dicePool, true); err != nil {
		return true
	}
	return false
}

func (g *Game) Rolled() bool {
	return g.rolled
}

func (g *Game) ScoreDiceHeld() (int, []int, error) {
	return score.Calculate(g.diceHeld, false)
}

func (g *Game) ScoreDicePool() (int, []int, error) {
	return score.Calculate(g.dicePool, true)
}

func (g *Game) DiceHeldCount() int {
	return len(g.diceHeld)
}

func (g *Game) DiceLockedCount() int {
	return len(g.diceLocked[len(g.diceLocked)-1])
}

func (g *Game) DicePoolHasFace(face int) bool {
	return g.dicePool.Contains(face)
}

func (g *Game) LockAllOfFace(face int) {
	c := 0
	for _, die := range g.dicePool {
		if die == face {
			c++
		}
	}
	for range c {
		g.HoldDie(face)
	}
}

func (g *Game) RenderDicePool(start, end int) string {
	return g.dicePool.Render(start, end)
}

func (g *Game) RenderDiceHeld(start, end int) string {
	return g.diceHeld.Render(start, end)
}

func (g *Game) RenderBankedDice() string {
	var bankedDie strings.Builder
	for _, diePool := range g.diceLocked {
		bankedDie.WriteString(diePool.RenderCharacters() + "\n")
	}
	return bankedDie.String()
}

func (g *Game) LockedScore() int {
	lockedScore := 0
	for _, diePool := range g.diceLocked {
		ls, _, _ := diePool.Score()
		lockedScore += ls
	}
	return lockedScore
}
