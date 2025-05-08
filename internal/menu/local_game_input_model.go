package menu

import (
	"strconv"
	"strings"

	"github.com/ascii-arcade/farkle/internal/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type localGameInputModel struct {
	width      int
	height     int
	focusIndex int
	inputs     []textinput.Model
	menuModel  menuModel
}

func newLocalGameInputModel(menuModel menuModel) localGameInputModel {
	m := localGameInputModel{
		inputs:    make([]textinput.Model, menuModel.numberOfPlayers),
		menuModel: menuModel,
		width:     menuModel.width,
		height:    menuModel.height,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 25
		t.Width = 25

		if i == 0 {
			t.Placeholder = "Player 1"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

			m.inputs[i] = t
			continue
		}

		t.Placeholder = "Player " + strconv.Itoa(i+1)
		m.inputs[i] = t
	}

	return m
}

func (m localGameInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m localGameInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				playerNames := make([]string, 0, len(m.inputs))
				for input := range m.inputs {
					playerNames = append(playerNames, m.inputs[input].Value())
				}

				tui.Run(playerNames, m.menuModel.debug)

				return m.menuModel, nil
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

func (m localGameInputModel) View() string {
	paneStyle := lipgloss.NewStyle().Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center)

	if m.height < 15 || m.width < 100 {
		return paneStyle.Render("Window too small, please resize to something larger.")
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
	rows = append(rows, "\n\n")
	submitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.focusIndex == len(m.inputs) {
		submitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center)
	}

	rows = append(rows, submitStyle.Render("Start Game"))

	inputPane := lipgloss.JoinVertical(lipgloss.Center, inputStyle.Render(strings.Join(rows, "\n")))
	return paneStyle.Render(inputPane)
}

func (m *localGameInputModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
