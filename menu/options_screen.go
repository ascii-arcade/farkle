package menu

import (
	"fmt"
	"strings"

	"github.com/ascii-arcade/farkle/board"
	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionScreen struct {
	model *Model
	style lipgloss.Style
}

func (m *Model) newOptionScreen() *optionScreen {
	return &optionScreen{
		model: m,
		style: m.style,
	}
}

func (s *optionScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *optionScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case tea.KeyMsg:
		switch {
		case keys.MenuEnglish.TriggeredBy(msg.String()):
			s.model.player.SetLanguage("en")
			return s.model, nil
		case keys.MenuSpanish.TriggeredBy(msg.String()):
			s.model.player.SetLanguage("es")
			return s.model, nil
		case keys.MenuEditProfile.TriggeredBy(msg.String()):
			if s.model.player.Visitor {
				return s.model, nil
			}
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newEditProfileScreen(),
				}
			}
		case keys.MenuStartNewGame.TriggeredBy(msg.String()):
			game, err := games.New(s.style)
			if err != nil {
				s.model.error = s.model.lang().Get("error", "game", err.Error())
				return s.model, nil
			}

			if err := game.AddPlayer(s.model.player); err != nil {
				s.model.error = s.model.lang().Get("error", "game", err.Error())
				return s.model, nil
			}
			return s.model, func() tea.Msg {
				return messages.SwitchViewMsg{
					Model: board.NewModel(s.style, s.model.width, s.model.height, s.model.player, game),
				}
			}
		case keys.MenuJoinGame.TriggeredBy(msg.String()):
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newJoinScreen(),
				}
			}
		}
	}

	return s.model, nil
}

func (s *optionScreen) View() string {
	var content strings.Builder
	content.WriteString(s.model.lang().Get("menu", "welcome") + "\n\n")
	content.WriteString(s.model.player.GetDisplayName(s.style) + "\n")
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_create"), keys.MenuStartNewGame.String(s.style)) + "\n")
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_join"), keys.MenuJoinGame.String(s.style)) + "\n")
	if !s.model.player.Visitor {
		content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_edit_profile"), keys.MenuEditProfile.String(s.style)) + "\n")
	}
	content.WriteString("\n\n")

	switch s.model.lang() {
	case language.Languages["en"]:
		content.WriteString(fmt.Sprintf(language.Languages["es"].Get("menu", "choose_language"), keys.MenuSpanish.String(s.style)))
	case language.Languages["es"]:
		content.WriteString(fmt.Sprintf(language.Languages["en"].Get("menu", "choose_language"), keys.MenuEnglish.String(s.style)))
	}

	content.WriteString(s.model.style.Foreground(colors.Faded).Render("\n\n" + config.Version))

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))

	return style.Render(content.String())
}
