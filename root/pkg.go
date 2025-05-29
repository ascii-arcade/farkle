package root

import (
	"errors"
	"log"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	wishTea "github.com/charmbracelet/wish/bubbletea"

	"github.com/ascii-arcade/farkle/game"
	gameModel "github.com/ascii-arcade/farkle/game_model"
	"github.com/ascii-arcade/farkle/menu"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/player"
)

type rootModel struct {
	active tea.Model
	menu   menu.Model
	game   gameModel.Model
}

func (m rootModel) Init() tea.Cmd {
	return m.active.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case messages.SwitchToMenu:
		m.active = m.menu
	case messages.SwitchToGame:
		m.active = m.game
		m.game.Init()
	}

	var cmd tea.Cmd
	newModel, cmd := m.active.Update(msg)
	m.active = newModel
	return m, cmd
}

func (m rootModel) View() string {
	return m.active.View()
}

func TeaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	renderer := wishTea.MakeRenderer(s)

	m := rootModel{
		game: gameModel.Model{
			Player:   player.New("Test"),
			Term:     pty.Term,
			Width:    pty.Window.Width,
			Height:   pty.Window.Height,
			Renderer: renderer,
		},
		menu: menu.Model{
			Term:     pty.Term,
			Width:    pty.Window.Width,
			Height:   pty.Window.Height,
			Renderer: renderer,
		},
	}
	m.active = m.menu

	err := m.newGame()
	if err != nil {
		log.Fatal("Could not create new game", "error", err)
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func (m *rootModel) newGame() error {
	code := newCode()
	game.Games[code] = game.New()
	return m.joinGame(code)
}

func (m *rootModel) joinGame(code string) error {
	updateCh := make(chan any)
	m.game.UpdateCh = updateCh
	m.game.GameCode = code

	state, exists := game.Games[code]
	if !exists {
		return errors.New("game does not exist")
	}

	state.AddClient(updateCh)
	state.Players[m.game.Player.Id] = &player.Player{
		Name:      m.game.Player.Name,
		TurnOrder: len(state.Players) + 1,
	}

	state.Refresh()

	return nil
}

func newCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b[:3]) + "-" + string(b[3:6])
}
