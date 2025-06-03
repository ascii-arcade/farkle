package menu

import (
	"strings"

	"github.com/ascii-arcade/farkle/board"
	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionScreen struct {
	model *Model
	style lipgloss.Style

	index         int
	gameCodeInput textinput.Model

	err string
}

func (m *Model) newOptionScreen() *optionScreen {
	gameRoomInput := textinput.New()
	gameRoomInput.Cursor.Style = m.style.Foreground(colors.Cursor)
	gameRoomInput.CharLimit = 7
	gameRoomInput.Width = 8
	gameRoomInput.Placeholder = "Game code"
	gameRoomInput.PromptStyle = m.style.Foreground(colors.Prompt)
	gameRoomInput.Focus()

	return &optionScreen{
		model:         m,
		style:         m.style,
		gameCodeInput: gameRoomInput,
	}
}

func (s *optionScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *optionScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyShiftTab:
			if s.index > 0 {
				s.index--
			}
		case tea.KeyDown, tea.KeyTab:
			if s.index < 1 {
				s.index++
			}
		}

		if keys.Submit.TriggeredBy(msg.String()) {
			if s.index == 0 {
				game := games.New(s.model.style)
				player := game.AddPlayer(true)
				gm := board.NewModel(s.model.style, s.model.Width, s.model.Height, player, game)
				return s.model, tea.Batch(
					func() tea.Msg {
						return messages.SwitchViewMsg{
							Model: gm,
						}
					},
					gm.Init(),
				)
			}

			if s.index == 1 {
				if len(s.gameCodeInput.Value()) < 7 {
					s.err = "A game code must be 7 characters long"
					s.index = 2
					return s.model, nil
				}
				game, ok := games.Get(s.gameCodeInput.Value())
				if !ok {
					s.err = "Game not found"
					s.gameCodeInput, cmd = s.gameCodeInput.Update(msg)
					s.gameCodeInput.SetValue("")
					return s.model, cmd
				}
				player := game.AddPlayer(false)
				gm := board.NewModel(s.model.style, s.model.Width, s.model.Height, player, game)
				return s.model, tea.Batch(
					func() tea.Msg {
						return messages.SwitchViewMsg{
							Model: gm,
						}
					},
					gm.Init(),
				)
			}
		}

		if msg.Type == tea.KeyCtrlQuestionMark {
			if len(s.gameCodeInput.Value()) == 4 {
				s.gameCodeInput.SetValue(s.gameCodeInput.Value()[:len(s.gameCodeInput.Value())-1])
			}
		}

		if msg.Type == tea.KeyEnter {
			s.err = ""
		}

		s.gameCodeInput, cmd = s.gameCodeInput.Update(msg)
		s.gameCodeInput.SetValue(strings.ToUpper(s.gameCodeInput.Value()))

		if len(s.gameCodeInput.Value()) == 3 && !strings.Contains(s.gameCodeInput.Value(), "-") {
			s.gameCodeInput.SetValue(s.gameCodeInput.Value() + "-")
			s.gameCodeInput.CursorEnd()
		}

		return s.model, cmd
	}

	return s.model, nil
}

func (s *optionScreen) View() string {
	if s.model.Width < 100 || s.model.Height < 20 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := s.model.style.Width(s.model.Width).Height(s.model.Height - 1).AlignVertical(lipgloss.Center)
	logoStyle := s.model.style.Foreground(colors.Logo).Margin(1, 2)
	titleStyle := s.model.style.Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := s.model.style.Foreground(colors.Default).AlignHorizontal(lipgloss.Left).Width(20)
	controlsStyle := s.model.style.Foreground(colors.Default).AlignHorizontal(lipgloss.Left).Width(s.model.Width / 2)
	errorsStyle := s.model.style.Foreground(colors.Error).AlignHorizontal(lipgloss.Right).Width(s.model.Width / 2)

	if config.GetDebug() {
		panelStyle = panelStyle.BorderForeground(colors.Debug).BorderStyle(lipgloss.ASCIIBorder()).Width(s.model.Width - 2).Height(s.model.Height - 3)
		logoStyle = logoStyle.BorderForeground(colors.Debug).BorderStyle(lipgloss.ASCIIBorder()).Margin(0, 1)
		menuStyle = menuStyle.BorderForeground(colors.Debug).BorderStyle(lipgloss.ASCIIBorder())
	}

	logoPanel := logoStyle.Render(logo)
	titlePanel := titleStyle.Render("Farkle")

	menu := make([]string, 0)
	style := s.model.style.Foreground(colors.Prompt)
	prefix := "   "
	if s.index == 0 {
		prefix = "-> "
	}
	menu = append(menu, style.Render(prefix+"New Online Game"))

	s.gameCodeInput.Prompt = "   "
	s.gameCodeInput.Blur()
	if s.index == 1 {
		s.gameCodeInput.Prompt = "-> "
		s.gameCodeInput.Focus()
	}
	sb := strings.Builder{}
	sb.WriteString(s.gameCodeInput.View())
	if len(s.gameCodeInput.Value()) == 7 && games.Exists(s.gameCodeInput.Value()) {
		sb.WriteString("\npress enter to join")
	}
	menu = append(menu, sb.String())

	menuContent := menuStyle.Render(strings.Join(menu, "\n"))

	menuJoin := lipgloss.JoinVertical(
		lipgloss.Center,
		titlePanel,
		menuContent,
	)

	logoWidth := lipgloss.Width(logoPanel)
	restOfPanelWidth := max(s.model.Width-logoWidth, 0)
	menuMargin := max((restOfPanelWidth/2)-lipgloss.Width(menuJoin), 0)

	menuPanel := s.model.style.MarginLeft(menuMargin).Align(lipgloss.Center, lipgloss.Center).Render(menuJoin)

	panel := panelStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Center,
		logoPanel,
		menuPanel,
	))

	controlsPane := lipgloss.JoinHorizontal(
		lipgloss.Left,
		controlsStyle.Render("ctrl+c to quit"),
		errorsStyle.Render(s.err),
	)

	return s.model.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		panel,
		controlsPane,
	))
}
