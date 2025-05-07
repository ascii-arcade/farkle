package menu

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/wsclient"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice struct {
	actionKeys []string
	action     func(m model) (tea.Model, tea.Cmd)
	render     func(m model) string
}

type model struct {
	width           int
	height          int
	cursor          int
	numberOfPlayers int
	choices         []menuChoice
	wsClient        *wsclient.Client
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tMsg(t)
	})
}

func Run(logger *slog.Logger) {
	if _, err := tea.NewProgram(new(logger)).Run(); err != nil {
		fmt.Println("Error starting program:", err)
	}
}

func new(logger *slog.Logger) *model {
	wsClient := wsclient.NewWsClient(logger, "ws://localhost:8080/ws")
	wsClient.Connect()

	return &model{
		cursor:          1,
		numberOfPlayers: 3,
		wsClient:        wsClient,
		choices: []menuChoice{
			{
				actionKeys: []string{"left", "right"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, nil
				},
				render: func(m model) string {
					return lipgloss.NewStyle().
						Foreground(lipgloss.Color("#fff")).
						Render(fmt.Sprintf("Number of Players: %d (←/→)", m.numberOfPlayers))
				},
			},
			{
				actionKeys: []string{"n"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return newPlayerNameInputModel(m), nil
				},
				render: func(m model) string {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("New Game (n)")
				},
			},
			{
				actionKeys: []string{"o"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, nil
				},
				render: func(m model) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))

					if !m.wsClient.IsConnected() {
						style = style.Foreground(lipgloss.Color("#ff0000"))
					}

					return style.Render("New Online Game (o)")
				},
			},
			{
				actionKeys: []string{"j"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, nil
				},
				render: func(m model) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))

					if !m.wsClient.IsConnected() {
						style = style.Foreground(lipgloss.Color("#ff0000"))
					}

					return style.Render("Join Game (j)")
				},
			},
			{
				actionKeys: []string{"e"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, tea.Quit
				},
				render: func(m model) string {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("Exit (e)")
				},
			},
		},
	}
}

type tMsg time.Time

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.cursor == 2 && !m.wsClient.IsConnected() {
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
			if m.cursor > 1 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case tea.KeyLeft:
			if m.numberOfPlayers > 2 {
				m.numberOfPlayers--
			}
		case tea.KeyRight:
			if m.numberOfPlayers < 6 {
				m.numberOfPlayers++
			}
		default:
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tMsg:
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.height < 15 || m.width < 100 {
		return "Window too small, please resize to something larger."
	}

	logoColor := lipgloss.Color("#0000ff")

	menuBaseStyle := lipgloss.NewStyle().Foreground(logoColor).BorderForeground(logoColor).Align(lipgloss.Center)

	logo := menuBaseStyle.Width(m.width / 3).AlignVertical(lipgloss.Center).Render(logo)
	title := menuBaseStyle.Border(lipgloss.NormalBorder()).Margin(1).Padding(1, 2).Align(lipgloss.Center, lipgloss.Center).Render("Farkle")
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
		title,
		menuBaseStyle.AlignHorizontal(lipgloss.Left).Render(strings.Join(menu, "\n")),
	)

	menuJoin = lipgloss.NewStyle().Width((m.width / 3) * 2).Height(m.height).AlignVertical(lipgloss.Center).Render(menuJoin)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		logo,
		menuJoin,
	)
}
