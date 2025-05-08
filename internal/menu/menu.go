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
	ticks           int

	logger *slog.Logger
	debug  bool
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
			continue
		}

		serverHealth = false
	}
}

func Run(logger *slog.Logger, debug bool) {
	if _, err := tea.NewProgram(new(logger, debug)).Run(); err != nil {
		fmt.Println("Error starting program:", err)
	}
}

func new(logger *slog.Logger, debug bool) *menuModel {
	wsClient = newWsClient(logger, "ws://localhost:8080/ws")

	return &menuModel{
		cursor:          1,
		numberOfPlayers: 3,
		debug:           debug,
		logger:          logger.With("component", "menu"),
		choices: []menuChoice{
			{
				actionKeys: []string{"left", "right"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					return m, nil
				},
				render: func(m menuModel) string {
					return lipgloss.NewStyle().
						Foreground(lipgloss.Color("#fff")).
						Render(fmt.Sprintf("Number of Players: %d (←/→)", m.numberOfPlayers))
				},
			},
			{
				actionKeys: []string{"n"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					return newLocalGameInputModel(m), nil
				},
				render: func(m menuModel) string {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("New Game (n)")
				},
			},
			{
				actionKeys: []string{"o"},
				action: func(m menuModel) (tea.Model, tea.Cmd) {
					if !serverHealth {
						return m, nil
					}

					return newOnlineGameInputModel(m), nil
				},
				render: func(m menuModel) string {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
					details := "o"

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
			m.ticks++
			if m.ticks%60 == 0 {
				m.ticks = 0
			}
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
	}

	return m, nil
}

func (m menuModel) View() string {
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

// func (m *menuModel) monitorMessages() {
// 	for msg := range wsClient.GetMessage() {
// 		switch msg.Channel {
// 		case server.ChannelLobby:
// 			switch msg.Type {
// 			case server.MessageTypeList:
// 				var lobbies []lobby.Lobby
// 				if err := json.Unmarshal(msg.Data, &lobbies); err != nil {
// 					m.logger.Error("Failed to unmarshal lobby list", "error", err)
// 					continue
// 				}
// 				m.logger.Debug("Received lobby list", "lobbies", lobbies)

// 			}
// 		}
// 	}
// }
