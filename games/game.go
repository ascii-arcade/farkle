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
	players []*Player
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

		player.Color = g.colors[len(g.players)%len(g.colors)]
		player.Name = utils.GenerateName(player.LanguagePreference.Lang)
		player.Score = 0

		if isHost {
			player.IsHost = true
		}

		player.OnDisconnect(func() {
			if !g.InProgress {
				g.RemovePlayer(player)
			}
		})

		g.players = append(g.players, player)
		return nil
	})
}

func (g *Game) RemovePlayer(player *Player) {
	g.withLock(func() {
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

		if g.GetPlayerCount(false) == 0 {
			delete(games, g.Code)
		}
	})
}

func (g *Game) GetPlayers() []*Player {
	return g.players
}

func (g *Game) Ready() bool {
	return len(g.players) >= 2 && !g.InProgress
}

func (g *Game) nextTurn() {
	score := 10000
	if config.GetDebug() {
		score = 1000
	}

	player := g.GetTurnPlayer()
	if player.Score >= score && !g.endGame {
		g.endGame = true
		g.log = append(g.log, player.StyledPlayerName(g.style)+" triggered end game!")
	}

	if g.IsGameOver() {
		winner := g.GetWinningPlayer()
		g.log = append(g.log, winner.StyledPlayerName(g.style)+" wins the game with a score of "+strconv.Itoa(winner.Score)+"!")
		return
	}

	if g.endGame && !player.PlayedLastTurn {
		player.PlayedLastTurn = true
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
		for _, p := range g.players {
			if p.IsHost {
				player = p
				break
			}
		}
	})
	return player
}

func (g *Game) GetTurnPlayer() *Player {
	for i, p := range g.players {
		if i == g.turn {
			return p
		}
	}
	return nil
}

func (g *Game) GetWinningPlayer() *Player {
	var player *Player
	if !g.endGame {
		return nil
	}

	for _, p := range g.players {
		if player == nil || p.Score > player.Score {
			player = p
		}
	}

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
	for _, p := range g.players {
		select {
		case p.UpdateChan <- struct{}{}:
		default:
		}
	}
}

func (s *Game) getPlayer(sess ssh.Session) (*Player, bool) {
	for _, p := range s.players {
		if p.Sess.User() == sess.User() {
			return p, true
		}
	}
	return nil, false
}

func (s *Game) GetDisconnectedPlayers() []*Player {
	var players []*Player
	s.withLock(func() {
		for _, p := range s.players {
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
	for _, p := range s.players {
		if includeDisconnected || p.connected {
			count++
		}
	}
	return count
}
