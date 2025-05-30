package menu

import (
	"strings"

	"github.com/ascii-arcade/farkle/config"
	gamemodel "github.com/ascii-arcade/farkle/game_model"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice struct {
	action func(Model, tea.Msg) (tea.Model, tea.Cmd)
	render func(Model, bool) string
	input  bool
}

type Model struct {
	width  int
	height int
	style  lipgloss.Style

	index         int
	gameCodeInput textinput.Model
	choices       []menuChoice

	err string
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func New(style lipgloss.Style, width, height int) *Model {
	gameRoomInput := textinput.New()
	gameRoomInput.Cursor.Style = style.Foreground(lipgloss.Color("205"))
	gameRoomInput.CharLimit = 7
	gameRoomInput.Width = 8
	gameRoomInput.Placeholder = "Game code"
	gameRoomInput.PromptStyle = style.Foreground(lipgloss.Color("#00ff00"))
	gameRoomInput.Focus()

	m := &Model{
		index:         0,
		gameCodeInput: gameRoomInput,
		width:         width,
		height:        height,
		style:         style,
	}

	m.choices = []menuChoice{
		{
			action: func(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
				game := games.New(m.style)
				player := game.AddPlayer(true)
				gm := gamemodel.NewModel(m.style, m.width, m.height, player, game)
				return m, tea.Batch(
					func() tea.Msg {
						return messages.SwitchViewMsg{
							Model: gm,
						}
					},
					gm.Init(),
				)
			},
			render: func(m Model, selected bool) string {
				style := m.style.Foreground(lipgloss.Color("#00ff00"))
				prefix := "   "
				if selected {
					prefix = "-> "
				}
				return style.Render(prefix + "New Online Game")
			},
		},
		{
			input: true,
			action: func(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
				var cmd tea.Cmd

				switch msg := msg.(type) {
				case tea.KeyMsg:
					if msg.Type == tea.KeyCtrlQuestionMark {
						if len(m.gameCodeInput.Value()) == 4 {
							m.gameCodeInput.SetValue(m.gameCodeInput.Value()[:len(m.gameCodeInput.Value())-1])
						}
					}
					if msg.Type == tea.KeyEnter {
						m.err = ""
						if len(m.gameCodeInput.Value()) < 7 {
							m.err = "A game code must be 7 characters long"
							m.index = 2
							return m, nil
						}
						game, ok := games.Get(m.gameCodeInput.Value())
						if !ok {
							m.err = "Game not found"
							m.gameCodeInput, cmd = m.gameCodeInput.Update(msg)
							m.gameCodeInput.SetValue("")
							return m, cmd
						}
						player := game.AddPlayer(false)
						gm := gamemodel.NewModel(m.style, m.width, m.height, player, game)
						return m, tea.Batch(
							func() tea.Msg {
								return messages.SwitchViewMsg{
									Model: gm,
								}
							},
							gm.Init(),
						)
					}
				}

				m.gameCodeInput, cmd = m.gameCodeInput.Update(msg)
				m.gameCodeInput.SetValue(strings.ToUpper(m.gameCodeInput.Value()))

				if len(m.gameCodeInput.Value()) == 3 && !strings.Contains(m.gameCodeInput.Value(), "-") {
					m.gameCodeInput.SetValue(m.gameCodeInput.Value() + "-")
					m.gameCodeInput.CursorEnd()
				}

				return m, cmd
			},
			render: func(m Model, selected bool) string {
				m.gameCodeInput.Prompt = "   "
				m.gameCodeInput.Blur()
				if selected {
					m.gameCodeInput.Prompt = "-> "
					m.gameCodeInput.Focus()
				}
				sb := strings.Builder{}
				sb.WriteString(m.gameCodeInput.View())
				if len(m.gameCodeInput.Value()) == 7 && games.Exists(m.gameCodeInput.Value()) {
					sb.WriteString(" (enter to join)")
				}
				return sb.String()
			},
		},
	}

	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
		case tea.KeyEnter:
			return m.choices[m.index].action(m, msg)
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyShiftTab:
			if m.index > 0 {
				m.index--
			}
		case tea.KeyDown, tea.KeyTab:
			if m.index < len(m.choices)-1 {
				m.index++
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	}

	for i, choice := range m.choices {
		if i == m.index && choice.input {
			return choice.action(m, msg)
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.height < 20 || m.width < 100 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := m.style.Width(m.width).Height(m.height - 1).AlignVertical(lipgloss.Center)
	logoStyle := m.style.Foreground(lipgloss.Color("#0000ff")).Margin(1, 2)
	titleStyle := m.style.Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := m.style.Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left)
	controlsStyle := m.style.Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width / 2)
	errorsStyle := m.style.Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.width / 2)

	if config.GetDebug() {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder()).Width(m.width - 2).Height(m.height - 3)
		logoStyle = logoStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder()).Margin(0, 1)
		menuStyle = menuStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder())
		controlsStyle = controlsStyle.Background(lipgloss.Color("#000066")).Foreground(lipgloss.Color("#ffffff"))
		errorsStyle = errorsStyle.Background(lipgloss.Color("#660000")).Foreground(lipgloss.Color("#ffffff"))
	}

	logoPanel := logoStyle.Render(logo)
	titlePanel := titleStyle.Render("Farkle")
	menu := make([]string, 0, len(m.choices))
	for i, choice := range m.choices {
		menu = append(menu, choice.render(m, m.index == i))
	}
	menuContent := menuStyle.Render(strings.Join(menu, "\n"))

	menuJoin := lipgloss.JoinVertical(
		lipgloss.Center,
		titlePanel,
		menuContent,
	)

	logoWidth := lipgloss.Width(logoPanel)
	restOfPanelWidth := max(m.width-logoWidth, 0)
	menuMargin := max((restOfPanelWidth/2)-lipgloss.Width(menuJoin), 0)

	menuPanel := m.style.MarginLeft(menuMargin).Align(lipgloss.Center, lipgloss.Center).Render(menuJoin)

	panel := panelStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Center,
		logoPanel,
		menuPanel,
	))

	controlsPane := lipgloss.JoinHorizontal(
		lipgloss.Left,
		controlsStyle.Render("ctrl+c to quit"),
		errorsStyle.Render(m.err),
	)

	return m.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		panel,
		controlsPane,
	))
}
