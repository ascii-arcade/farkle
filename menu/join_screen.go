package menu

import (
	"errors"
	"strings"

	"github.com/ascii-arcade/farkle/board"
	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type joinScreen struct {
	model *Model
	style lipgloss.Style

	gameCodeInput textinput.Model
}

func (m *Model) newJoinScreen() *joinScreen {
	s := &joinScreen{
		model:         m,
		style:         m.style,
		gameCodeInput: textinput.New(),
	}
	s.gameCodeInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	s.gameCodeInput.CharLimit = 7
	s.gameCodeInput.Width = 8
	s.gameCodeInput.Placeholder = m.lang().Get("menu", "join", "game_code_placeholder")
	s.gameCodeInput.PromptStyle = m.style.Foreground(lipgloss.Color("#00ff00"))
	s.gameCodeInput.Focus()
	return s
}

func (s *joinScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *joinScreen) Update(msg tea.Msg) (any, tea.Cmd) {
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
		if keys.Submit.TriggeredBy(msg.String()) {
			if len(s.gameCodeInput.Value()) == 7 {
				code := strings.ToUpper(s.gameCodeInput.Value())
				game, err := games.GetOpenGame(code)
				if err != nil && (errors.Is(err, games.ErrGameAlreadyInProgress) && game.HasPlayer(s.model.player)) {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}

				if err := game.AddPlayer(s.model.player, false); err != nil {
					s.model.error = s.model.lang().Get("error", "game", err.Error())
					return s.model, nil
				}

				return s.model, func() tea.Msg {
					return messages.SwitchViewMsg{
						Model: board.NewModel(s.style, s.model.width, s.model.height, s.model.player, game),
					}
				}
			}
		}

		if msg.Type == tea.KeyCtrlQuestionMark {
			if len(s.gameCodeInput.Value()) == 4 {
				s.gameCodeInput.SetValue(s.gameCodeInput.Value()[:len(s.gameCodeInput.Value())-1])
			}
		}

		if msg.Type == tea.KeyEnter {
			s.model.error = ""
		}
	}

	s.gameCodeInput, cmd = s.gameCodeInput.Update(msg)
	s.gameCodeInput.SetValue(strings.ToUpper(s.gameCodeInput.Value()))

	if len(s.gameCodeInput.Value()) == 3 && !strings.Contains(s.gameCodeInput.Value(), "-") {
		s.gameCodeInput.SetValue(s.gameCodeInput.Value() + "-")
		s.gameCodeInput.CursorEnd()
	}
	return s.model, cmd
}

func (s *joinScreen) View() string {
	errorMessage := s.model.error

	var content strings.Builder
	content.WriteString(s.model.lang().Get("menu", "join", "enter_code") + "\n\n")
	content.WriteString(s.gameCodeInput.View() + "\n\n")
	content.WriteString(s.style.Foreground(colors.Error).Render(errorMessage))
	content.WriteString("\n\n")

	return content.String()
}
