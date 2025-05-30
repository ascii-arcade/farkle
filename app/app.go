package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/ascii-arcade/farkle/messages"
	splashscreen "github.com/ascii-arcade/farkle/splash_screen"
)

type rootModel struct {
	active tea.Model
	sess   ssh.Session
}

func (m rootModel) Init() tea.Cmd {
	return m.active.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchViewMsg:
		m.active = msg.Model
		initcmd := m.active.Init()
		return m, initcmd
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	return m, cmd
}

func (m rootModel) View() string {
	return m.active.View()
}

func TeaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		return nil, nil
	}
	return rootModel{
		sess:   s,
		active: splashscreen.NewModel(bubbletea.MakeRenderer(s).NewStyle(), pty.Window.Width, pty.Window.Height),
	}, []tea.ProgramOption{tea.WithAltScreen()}
}

// func (m *rootModel) newGame() error {
// 	code := newCode()
// 	game.Games[code] = game.New()
// 	return m.joinGame(code)
// }

// func (m *rootModel) joinGame(code string) error {
// 	updateCh := make(chan any)
// 	m.game.UpdateCh = updateCh
// 	m.game.GameCode = code

// 	state, exists := game.Games[code]
// 	if !exists {
// 		return errors.New("game does not exist")
// 	}

// 	state.AddClient(updateCh)
// 	state.Players[m.game.Player.Id] = &player.Player{
// 		Name:      m.game.Player.Name,
// 		TurnOrder: len(state.Players) + 1,
// 	}

// 	state.Refresh()

// 	return nil
// }

// func newCode() string {
// 	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
// 	b := make([]byte, 6)
// 	for i := range b {
// 		b[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return string(b[:3]) + "-" + string(b[3:6])
// }
