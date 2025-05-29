package menu

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/config"
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
	Term     string
	Width    int
	Height   int
	Renderer *lipgloss.Renderer

	isJoining bool
	gameCode  string

	index           int
	playerNameInput textinput.Model
	gameCodeInput   textinput.Model
	choices         []menuChoice
	serverHealthy   bool

	err string

	logger *slog.Logger
}

type serverHealthMsg bool

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), serverHealth(false))
}

func New(logger *slog.Logger) *Model {
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

	m := &Model{
		index:           0,
		playerNameInput: playerNameInput,
		gameCodeInput:   gameRoomInput,
		logger:          logger,
	}

	m.choices = []menuChoice{
		{
			input: true,
			action: func(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
				var cmd tea.Cmd
				m.playerNameInput, cmd = m.playerNameInput.Update(msg)
				return m, cmd
			},
			render: func(m Model, selected bool) string {
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
			action: func(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
				return m, nil
			},
			render: func(m Model, selected bool) string {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
				prefix := "   "
				if selected {
					prefix = "-> "
				}
				if !m.serverHealthy {
					style = style.Foreground(lipgloss.Color("#ff0000"))
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
						if len(m.playerNameInput.Value()) < 3 {
							m.err = "A player name must be at least 3 characters long"
							m.index = 0
							return m, nil
						}
						if len(m.gameCodeInput.Value()) < 7 {
							m.err = "A game code must be 7 characters long"
							m.index = 2
							return m, nil
						}
						return m, nil
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
				return m.gameCodeInput.View()
			},
		},
		{
			action: func(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
				return m, tea.Quit
			},
			render: func(m Model, selected bool) string {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
				prefix := "   "
				if selected {
					prefix = "-> "
				}
				return style.Render(prefix + "Exit")
			},
		},
	}

	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case serverHealthMsg:
		m.serverHealthy = bool(msg)
		return m, serverHealth(m.serverHealthy)
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
		m.Height = msg.Height
		m.Width = msg.Width
	}

	for i, choice := range m.choices {
		if i == m.index && choice.input {
			return choice.action(m, msg)
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.Height < 20 || m.Width < 100 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := lipgloss.NewStyle().Width(m.Width).Height(m.Height - 1).AlignVertical(lipgloss.Center)
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff")).Margin(1, 2)
	titleStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.Width / 2)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.Width / 2)

	if config.GetDebug() {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder()).Width(m.Width - 2).Height(m.Height - 3)
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

	if !m.serverHealthy {
		menuContent = menuStyle.Render("Connecting to server...")
	}

	menuJoin := lipgloss.JoinVertical(
		lipgloss.Center,
		titlePanel,
		menuContent,
	)

	logoWidth := lipgloss.Width(logoPanel)
	restOfPanelWidth := max(m.Width-logoWidth, 0)
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

func serverHealth(healthy bool) tea.Cmd {
	return func() tea.Msg {
		if healthy {
			time.Sleep(1 * time.Second)
		}

		client := http.Client{}
		scheme := "http"
		if config.GetSecure() {
			scheme = "https"
		}

		res, err := client.Get(fmt.Sprintf("%s://%s:%s/health", scheme, config.GetServerURL(), config.GetServerPort()))
		if err != nil {
			return serverHealthMsg(false)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return serverHealthMsg(false)
		}

		return serverHealthMsg(true)
	}
}
