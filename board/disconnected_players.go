package board

import (
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type disconnectedPlayersScreen struct {
	model *Model
}

func (s *disconnectedPlayersScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *disconnectedPlayersScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	return s.model, nil
}

func (s *disconnectedPlayersScreen) View() string {
	if s.model.width < 20 {
		return s.model.lang().Get("error", "window_too_narrow")
	}
	if s.model.height < 10 {
		return s.model.lang().Get("error", "window_too_short")
	}

	style := s.model.style.Width(s.model.width).Height(s.model.height)
	paneStyle := style.Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(s.model.player.Color)

	content := "Disconnected Players:\n"
	for _, player := range s.model.game.GetDisconnectedPlayers() {
		content += "- " + player.Name + "\n"
	}

	return paneStyle.Render(content)
}
