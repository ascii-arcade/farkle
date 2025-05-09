package menu

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice struct {
	actionKeys []string
	action     func(m menuModel) (tea.Model, tea.Cmd)
	render     func(m menuModel) string
}

type menuModel struct {
	width           int
	height          int
	cursor          int
	numberOfPlayers int
	choices         []menuChoice

	logger *slog.Logger
}

func (m menuModel) Init() tea.Cmd {
	go m.checkHealth()
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

func (m *menuModel) checkHealth() {
	c := http.Client{}
	for {
		resp, err := c.Get("http://localhost:8080/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			serverHealth = true
			goto CONTINUE
		}
		serverHealth = false

	CONTINUE:
		time.Sleep(5 * time.Second)
	}
}

func Run(logger *slog.Logger, debug bool) {
	if _, err := tea.NewProgram(newMenu(logger, debug)).Run(); err != nil {
		fmt.Println("Error starting program:", err)
	}
}

func newMenu(logger *slog.Logger, d bool) *menuModel {
	debug = d
	return &menuModel{
		cursor:          0,
		numberOfPlayers: 3,
		logger:          logger.With("component", "menu"),
		choices: []menuChoice{
			{
				actionKeys: []string{"n"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					if !serverHealth {
						return m, nil
					}

					nm := newLobbyInputModel(m)

					return nm, nm.Init()
				},
				render: func(m menuModel) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					details := "n"

					if !serverHealth {
						style = style.Foreground(lipgloss.Color("#ff0000"))
						details = "connecting..."
					}

					return style.Render(fmt.Sprintf("New Online Game (%s)", details))
				},
			},
			{
				actionKeys: []string{"j"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					if !serverHealth {
						return m, nil
					}

					return newJoinGameModel(m), nil
				},
				render: func(m menuModel) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					details := "j"

					if !serverHealth {
						style = style.Foreground(lipgloss.Color("#ff0000"))
						details = "connecting..."
					}

					return style.Render(fmt.Sprintf("Join Game (%s)", details))
				},
			},
			{
				actionKeys: []string{"e"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					return m, tea.Quit
				},
				render: func(m menuModel) string {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("Exit (e)")
				},
			},
		},
	}
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tick:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			m.logger.Debug("Tick")
			return tick(t)
		})
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.cursor == 2 && !serverHealth {
				return m, nil
			}
			return m.choices[m.cursor].action(m)
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "e":
				return m, tea.Quit
			default:
				for _, choice := range m.choices {
					if slices.Contains(choice.actionKeys, string(msg.Runes)) {
						return choice.action(m)
					}
				}
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		default:
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, nil
}

func (m menuModel) View() string {
	if m.height < 15 || m.width < 100 {
		return "Window too small, please resize to something larger."
	}

	panelStyle := lipgloss.NewStyle().Width(m.width).Height(m.height).AlignVertical(lipgloss.Center)
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff")).Margin(1, 2)
	titleStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)
	menuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Left)

	if debug {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder()).Width(m.width-3).Height(m.height-2).Margin(0, 1)
		logoStyle = logoStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder()).Margin(0, 1)
		menuStyle = menuStyle.BorderForeground(lipgloss.Color("#ff0000")).BorderStyle(lipgloss.ASCIIBorder())
	}

	logoPanel := logoStyle.Render(logo)
	titlePanel := titleStyle.Render("Farkle")
	menu := make([]string, 0, 4)
	for i, choice := range m.choices {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
		prefix := "   "

		if i == m.cursor {
			style = style.Foreground(lipgloss.Color("#00ff00"))
			prefix = "-> "
		}

		menu = append(menu, style.Render(prefix+choice.render(m)))
	}

	menuJoin := lipgloss.JoinVertical(
		lipgloss.Center,
		titlePanel,
		menuStyle.Render(strings.Join(menu, "\n")),
	)

	logoWidth := lipgloss.Width(logoPanel)
	restOfPanelWidth := max(m.width-logoWidth, 0)
	menuMargin := max((restOfPanelWidth/2)-lipgloss.Width(menuJoin), 0)

	menuPanel := lipgloss.NewStyle().MarginLeft(menuMargin).Align(lipgloss.Center, lipgloss.Center).Render(menuJoin)

	return panelStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Center,
		logoPanel,
		menuPanel,
	))
}
