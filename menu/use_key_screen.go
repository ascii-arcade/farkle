package menu

import (
	"strings"

	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type useKeyScreen struct {
	model *Model
	style lipgloss.Style

	keyInput textarea.Model
}

func (m *Model) newUseKeyScreen() *useKeyScreen {
	s := &useKeyScreen{
		model:    m,
		style:    m.style,
		keyInput: textarea.New(),
	}
	s.keyInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	s.keyInput.Placeholder = ""
	s.keyInput.SetWidth(70)
	s.keyInput.SetHeight(10)
	s.keyInput.CharLimit = 600
	s.keyInput.Focus()
	return s
}

func (s *useKeyScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *useKeyScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case cursor.BlinkMsg:
		return s.model, nil

	case tea.KeyMsg:
		s.model.error = ""
		if keys.PreviousScreen.TriggeredBy(msg.String()) {
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newOptionScreen(),
				}
			}
		}
		// if keys.Submit.TriggeredBy(msg.String()) {
		// 	if utils.ValidPublicKey(s.keyInput.Value()) {
		// 		s.model.player.AddPubKey(s.keyInput.Value())
		// 		if err := s.model.player.Save(); err != nil {
		// 			s.model.error = "Internal Error"
		// 			slog.Error("error saving player with public key", "error", err)
		// 			return s.model, cmd
		// 		}

		// 		return s.model, func() tea.Msg {
		// 			return messages.SwitchScreenMsg{
		// 				Screen: s.model.newOptionScreen(),
		// 			}
		// 		}
		// 	}

		// 	s.model.error = s.model.lang().Get("menu", "add_key", "invalid_key")
		// 	return s.model, cmd
		// }

		if msg.Type == tea.KeyEnter {
			s.model.error = ""
		}
	}

	s.keyInput, cmd = s.keyInput.Update(msg)
	return s.model, cmd
}

func (s *useKeyScreen) View() string {
	errorMessage := s.model.error

	var content strings.Builder
	content.WriteString(s.model.lang().Get("menu", "use_key", "enter_key") + "\n\n")
	content.WriteString(s.keyInput.View() + "\n\n")
	content.WriteString(s.style.Foreground(colors.Error).Render(errorMessage))
	content.WriteString("\n\n")

	return content.String()
}
