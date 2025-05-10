package menu

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice struct {
	action func(menuModel, tea.Msg) (tea.Model, tea.Cmd)
	render func(menuModel, bool) string
	input  bool
}

type menuModel struct {
	width           int
	height          int
	index           int
	playerNameInput textinput.Model
	gameCodeInput   textinput.Model
	choices         []menuChoice

	err string
}

func (m menuModel) Init() tea.Cmd {
	if wsClient != nil {
		close(wsClient.disconnect)
		wsClient = nil
	}

	sizeCmd := tea.WindowSize()
	tickCmd := tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tick(t)
	})

	return tea.Batch(sizeCmd, tickCmd)
}

func (m *menuModel) checkHealth() {
	c := http.Client{}
	// for {
	resp, err := c.Get("http://localhost:8080/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		serverHealth = true
		return
	}
	serverHealth = false

	// CONTINUE:
	// 	time.Sleep(5 * time.Second)
	// }
}

func Run(loggerIn *slog.Logger, debugIn bool) {
	logger = loggerIn
	debug = debugIn
	if _, err := tea.NewProgram(newMenu()).Run(); err != nil {
		fmt.Println("Error starting program:", err)
	}
}

func newMenu() *menuModel {
	playerNameInput := textinput.New()
	playerNameInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	playerNameInput.CharLimit = 25
	playerNameInput.Width = 25
	playerNameInput.Placeholder = "Your name"
	playerNameInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	playerNameInput.Focus()

	gameRoomInput := textinput.New()
	gameRoomInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	gameRoomInput.CharLimit = 7
	gameRoomInput.Width = 25
	gameRoomInput.Placeholder = "Game code"
	gameRoomInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	gameRoomInput.Focus()

	return &menuModel{
		index:           0,
		playerNameInput: playerNameInput,
		gameCodeInput:   gameRoomInput,
		choices: []menuChoice{
			{
				input: true,
				action: func(m menuModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					var cmd tea.Cmd
					m.playerNameInput, cmd = m.playerNameInput.Update(msg)
					return m, cmd
				},
				render: func(m menuModel, selected bool) string {
					m.playerNameInput.Prompt = "   "
					m.playerNameInput.Blur()
					if selected {
						m.playerNameInput.Prompt = "-> "
						m.playerNameInput.Focus()
					}
					return m.playerNameInput.View()
				},
			},
			{
				action: func(m menuModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					if !serverHealth {
						return m, nil
					}
					return newLobbyModel(m.playerNameInput.Value(), "", false)
				},
				render: func(m menuModel, selected bool) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					prefix := "   "

					if selected {
						style = style.Foreground(lipgloss.Color("#00ff00"))
						prefix = "-> "
					}

					if !serverHealth {
						style = style.Foreground(lipgloss.Color("#ff0000"))
					}

					return style.Render(prefix + "New Online Game")
				},
			},
			{
				input: true,
				action: func(m menuModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					var cmd tea.Cmd

					switch msg := msg.(type) {
					case tea.KeyMsg:
						if msg.Type == tea.KeyCtrlQuestionMark {
							if len(m.gameCodeInput.Value()) == 4 {
								m.gameCodeInput.SetValue(m.gameCodeInput.Value()[:len(m.gameCodeInput.Value())-1])
							}
						}
						if msg.Type == tea.KeyEnter {
							if len(m.playerNameInput.Value()) < 3 {
								m.err = "A player name must be at least 3 characters long"
								m.index = 0
								return m, nil
							}

							if len(m.gameCodeInput.Value()) < 7 {
								m.err = "A game code must be 7 characters long"
								m.index = 3
								return m, nil
							}

							return newLobbyModel(m.playerNameInput.Value(), m.gameCodeInput.Value(), true)
						}
					}

					m.gameCodeInput, cmd = m.gameCodeInput.Update(msg)

					if len(m.gameCodeInput.Value()) == 3 {
						m.gameCodeInput.SetValue(m.gameCodeInput.Value() + "-")
						m.gameCodeInput.CursorEnd()
					}

					return m, cmd
				},
				render: func(m menuModel, selected bool) string {
					m.gameCodeInput.Prompt = "   "
					m.gameCodeInput.Blur()

					if selected {
						m.gameCodeInput.Prompt = "-> "
						m.gameCodeInput.Focus()
					}

					return m.gameCodeInput.View()
				},
			},
			{
				action: func(m menuModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					return m, tea.Quit
				},
				render: func(m menuModel, selected bool) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					prefix := "   "

					if selected {
						style = style.Foreground(lipgloss.Color("#00ff00"))
						prefix = "-> "
					}
					return style.Render(prefix + "Exit")
				},
			},
		},
	}
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tick:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			m.checkHealth()
			return tick(t)
		})
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
		case tea.KeyEnter:
			return m.choices[m.index].action(m, msg)
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.index > 0 {
				m.index--
			}
		case tea.KeyDown:
			if m.index < len(m.choices)-1 {
				m.index++
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	for i, choice := range m.choices {
		if i == m.index && choice.input {
			return choice.action(m, msg)
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	if m.height < 15 || m.width < 100 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := lipgloss.NewStyle().Width(m.width).Height(m.height - 1).AlignVertical(lipgloss.Center)
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff")).Margin(1, 2)
	titleStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width / 2)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.width / 2)

	if debug {
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

	if !serverHealth {
		menuContent = menuStyle.Render("Connecting to server...")
	}

	menuJoin := lipgloss.JoinVertical(
		lipgloss.Center,
		titlePanel,
		menuContent,
	)

	logoWidth := lipgloss.Width(logoPanel)
	restOfPanelWidth := max(m.width-logoWidth, 0)
	menuMargin := max((restOfPanelWidth/2)-lipgloss.Width(menuJoin), 0)

	menuPanel := lipgloss.NewStyle().MarginLeft(menuMargin).Align(lipgloss.Center, lipgloss.Center).Render(menuJoin)

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

	return lipgloss.JoinVertical(
		lipgloss.Center,
		panel,
		controlsPane,
	)
}
