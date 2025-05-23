package menu

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/client/eventloop"
	"github.com/ascii-arcade/farkle/client/lobby"
	"github.com/ascii-arcade/farkle/client/networkmanager"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/lobbies"
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
	serverHealthy   bool

	err string

	logger *slog.Logger
}

type serverHealthMsg bool

func (m menuModel) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), serverHealth(false))
}

func New() *menuModel {
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

	m := &menuModel{
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
					if !m.serverHealthy {
						return m, nil
					}
					client := http.Client{}
					res, err := client.Post("http://localhost:8080/lobbies", "application/json", nil)
					if err != nil {
						m.logger.Debug("FAILED")
						return m, nil
					}
					body, err := io.ReadAll(res.Body)
					if err != nil {
						m.logger.Debug("FAILED")
						return m, nil
					}
					var l *lobbies.Lobby
					if err := json.Unmarshal(body, &l); err != nil {
						m.logger.Debug("FAILED")
						return m, nil
					}

					nm, err := networkmanager.NewNetworkManager("ws://" + config.GetServerURL() + ":" + config.GetServerPort() + "/ws/" + l.Code + "?name=" + m.playerNameInput.Value())
					if err != nil {
						return m, nil
					}

					p := tea.NewProgram(lobby.New(nm), tea.WithAltScreen())

					eventLoop := eventloop.New(nm.Incoming, p)
					eventLoop.Start()

					if _, err := p.Run(); err != nil {
						m.logger.Error("Error running client", "error", err)
					}

					return m, nil
				},
				render: func(m menuModel, selected bool) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					prefix := "   "

					if selected {
						style = style.Foreground(lipgloss.Color("#00ff00"))
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

							nm, err := networkmanager.NewNetworkManager("ws://" + config.GetServerURL() + ":" + config.GetServerPort() + "/ws/" + m.gameCodeInput.Value() + "?name=" + m.playerNameInput.Value())
							if err != nil {
								return m, nil
							}

							p := tea.NewProgram(lobby.New(nm), tea.WithAltScreen())

							eventLoop := eventloop.New(nm.Incoming, p)
							eventLoop.Start()

							if _, err := p.Run(); err != nil {
								m.logger.Error("Error running client", "error", err)
							}

							return m, nil
						}
					}

					m.gameCodeInput, cmd = m.gameCodeInput.Update(msg)

					m.gameCodeInput.SetValue(strings.ToUpper(m.gameCodeInput.Value()))

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

	return m
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	if m.height < 20 || m.width < 100 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := lipgloss.NewStyle().Width(m.width).Height(m.height - 1).AlignVertical(lipgloss.Center)
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff")).Margin(1, 2)
	titleStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left).Width(m.width / 2)
	errorsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).AlignHorizontal(lipgloss.Right).Width(m.width / 2)

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

	if !m.serverHealthy {
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

func serverHealth(healthy bool) tea.Cmd {
	return func() tea.Msg {
		if healthy {
			time.Sleep(1 * time.Second)
		}

		client := http.Client{}
		res, err := client.Get("http://localhost:8080/health")
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
