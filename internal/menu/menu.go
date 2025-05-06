package menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/tui"
	"github.com/ascii-arcade/farkle/internal/wsclient"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuChoice struct {
	s         string
	shortKeys []string
	action    func(m model) (tea.Model, tea.Cmd)
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

func Run() {
	if _, err := tea.NewProgram(new()).Run(); err != nil {
		fmt.Println("Error starting program:", err)
	}
}

func new() *model {
	return &model{
		cursor:          1,
		numberOfPlayers: 3,
		wsClient:        wsclient.NewWsClient(nil, "ws://localhost:8080/ws"),
		choices: []menuChoice{
			{
				s:         "Number of Players",
				shortKeys: []string{"←", "→"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, nil
				},
			},
			{
				s:         "New Game",
				shortKeys: []string{"n"},
				action: func(m model) (tea.Model, tea.Cmd) {
					tui.Run([]string{"Test", "Test2"})
					return m, nil
				},
			},
			{
				s:         "Join Game",
				shortKeys: []string{"j"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, nil
				},
			},
			{
				s:         "Exit",
				shortKeys: []string{"e"},
				action: func(m model) (tea.Model, tea.Cmd) {
					return m, tea.Quit
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
			if m.cursor == 2 && !m.wsClient.Connected() {
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
					if strings.Contains(choice.s, string(msg.Runes)) {
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
			if m.numberOfPlayers < 5 {
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

	logo := menuBaseStyle.Width(m.width / 3).AlignVertical(lipgloss.Center).Render("")
	title := menuBaseStyle.Border(lipgloss.NormalBorder()).Margin(1).Padding(1, 2).Align(lipgloss.Center, lipgloss.Center).Render("Farkle")
	menu := make([]string, 0, 4)
	for i, choice := range m.choices {
		sk := strings.Join(choice.shortKeys, "/")
		if i == 0 {
			menu = append(menu, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Render(fmt.Sprintf("Number of Players: %d (%s)", m.numberOfPlayers, sk)))
			continue
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
		prefix := "   "

		if i == m.cursor {
			style = style.Foreground(lipgloss.Color("#00ff00"))
			prefix = "-> "
		}

		if i == 2 && !m.wsClient.Connected() {
			style = style.Foreground(lipgloss.Color("#ff0000"))
			menu = append(menu, style.Render(fmt.Sprintf("%s%s (connecting)", prefix, choice.s)))
		} else {
			menu = append(menu, style.Render(fmt.Sprintf("%s%s (%s)", prefix, choice.s, sk)))
		}
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
