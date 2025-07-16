package games

import (
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/score"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
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
	colors  []lipgloss.Color
	turn    int
	log     []string
	players map[*Player]*PlayerData
	style   lipgloss.Style
	mu      sync.Mutex
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

func (g *Game) AddPlayer(player *Player, isHost bool) error {
	return g.withErrLock(func() error {
		if _, ok := g.getPlayer(player.Sess); ok {
			return nil
		}

		if g.InProgress {
			return ErrGameAlreadyInProgress
		}

		playerData := &PlayerData{
			Name:      utils.GenerateName(player.LanguagePreference.Lang),
			Score:     0,
			Color:     g.colors[len(g.players)%len(g.colors)],
			turnOrder: len(g.players),
		}

		if isHost {
			playerData.IsHost = true
		}

		player.OnDisconnect(func() {
			if !g.InProgress {
				g.RemovePlayer(player)
			}
		})

		g.players[player] = playerData
		return nil
	})
}

func (g *Game) RemovePlayer(player *Player) {
	g.withLock(func() {
		delete(g.players, player)

		if len(g.players) > 0 && g.players[player].IsHost {
			for _, pd := range g.players {
				if !pd.IsHost {
					pd.IsHost = true
					break
				}
			}
		}

		if len(g.players) == 1 && g.InProgress {
			g.InProgress = false
		}

		if g.GetPlayerCount(false) == 0 {
			delete(games, g.Code)
		}
	})
}

func (g *Game) GetPlayers() []*Player {
	players := make([]*Player, 0, len(g.players))
	for p := range g.players {
		players = append(players, p)
	}
	slices.SortFunc(players, func(a, b *Player) int {
		return g.players[a].turnOrder - g.players[b].turnOrder
	})
	return players
}

func (g *Game) Ready() bool {
	return len(g.players) >= 2 && !g.InProgress
}

func (g *Game) GetTurnPlayer() *Player {
	for p, pd := range g.players {
		if pd.turnOrder == g.turn {
			return p
		}
	}
	return nil
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
		return
	}

	if g.endGame && !playerData.PlayedLastTurn {
		playerData.PlayedLastTurn = true
	}

	g.turn++
	if g.turn >= len(g.players) {
		g.turn = 0
	}
	g.Rolled = false
	g.Busted = false
	g.DiceLocked = []dice.DicePool{}
	g.DicePool = dice.NewDicePool(6)
	g.DiceHeld = dice.NewDicePool(0)
	g.DiceLocked = make([]dice.DicePool, 0)
	g.FirstRoll = false
}

func (g *Game) GetHost() *Player {
	var player *Player
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

func (g *Game) GetWinningPlayer() *Player {
	var winningPlayer *Player
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

func (g *Game) IsTurn(p *Player) bool {
	return g.players[g.GetTurnPlayer()].Name == g.players[p].Name
}

func (g *Game) PlayerScores() string {
	scores := make([]string, 0, len(g.players))

	for _, p := range g.GetPlayers() {
		pd := g.GetPlayerData(p)
		isCurrentPlayer := g.turn == pd.turnOrder
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

func (g *Game) busted() bool {
	if _, _, err := score.Calculate(g.DicePool, true); err != nil {
		return true
	}
	return false
}

func (g *Game) Refresh() {
	for p, _ := range g.players {
		select {
		case p.UpdateChan <- struct{}{}:
		default:
		}
	}
}

func (s *Game) getPlayer(sess ssh.Session) (*Player, bool) {
	for p, _ := range s.players {
		if p.Sess.User() == sess.User() {
			return p, true
		}
	}
	return nil, false
}

func (s *Game) GetDisconnectedPlayers() []*Player {
	var players []*Player
	s.withLock(func() {
		for p, _ := range s.players {
			if !p.connected {
				players = append(players, p)
			}
		}
	})
	return players
}

func (s *Game) HasPlayer(player *Player) bool {
	_, exists := s.getPlayer(player.Sess)
	return exists
}

func (s *Game) GetPlayerCount(includeDisconnected bool) int {
	count := 0
	for p, _ := range s.players {
		if includeDisconnected || p.connected {
			count++
		}
	}
	return count
}

func (s *Game) GetPlayerData(player *Player) *PlayerData {
	return s.players[player]
}
