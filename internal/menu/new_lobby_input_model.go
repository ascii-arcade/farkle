package menu

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyInputModel struct {
	width      int
	height     int
	focusIndex int
	inputs     []textinput.Model
	menuModel  menuModel
}

func newLobbyInputModel(menuModel menuModel) lobbyInputModel {
	m := lobbyInputModel{
		inputs:    make([]textinput.Model, 0),
		menuModel: menuModel,
		width:     menuModel.width,
		height:    menuModel.height,
	}

	playerNameInput := textinput.New()
	playerNameInput.Width = 25
	playerNameInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	playerNameInput.CharLimit = 25
	playerNameInput.Placeholder = "Your name"
	playerNameInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	playerNameInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	playerNameInput.Focus()

	lobbyNameInput := textinput.New()
	lobbyNameInput.Width = 25
	lobbyNameInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	lobbyNameInput.CharLimit = 25
	lobbyNameInput.Placeholder = "Lobby name"
	lobbyNameInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	lobbyNameInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m.inputs = append(m.inputs, playerNameInput)
	m.inputs = append(m.inputs, lobbyNameInput)

	return m
}

func (m lobbyInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m lobbyInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m.menuModel, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tick(t)
			})
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				playerName := m.inputs[0].Value()
				lobbyName := m.inputs[1].Value()

				if playerName == "" || lobbyName == "" {
					return m, nil
				}
				nm := newLobbyModel(m.menuModel, lobbyName, playerName)

				return nm, nm.Init()
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			}
			if s == "down" || s == "tab" {
				m.focusIndex++
			}

			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			} else if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					continue
				}

				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m lobbyInputModel) View() string {
	paneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center)

	if m.height < 15 || m.width < 100 {
		return paneStyle.Render("Window too small, please resize to something larger.")
	}

	if debug {
		paneStyle = paneStyle.BorderStyle(lipgloss.ASCIIBorder()).BorderForeground(lipgloss.Color("#0000ff")).Width(m.width - 2).Height(m.height - 2)
	}

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Border(lipgloss.NormalBorder()).
		Margin(1).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)

	rows := make([]string, 0)
	for i := range m.inputs {
		rows = append(rows, m.inputs[i].View())
	}
	submitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1)
	if m.focusIndex == len(m.inputs) {
		submitStyle = submitStyle.
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center)
	}

	rows = append(rows, submitStyle.Render("Create Lobby"))

	inputPane := lipgloss.JoinVertical(lipgloss.Center, inputStyle.Render(strings.Join(rows, "\n")))
	return paneStyle.Render(inputPane)
}

func (m *lobbyInputModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
