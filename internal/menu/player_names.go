package menu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ascii-arcade/farkle/internal/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type playerNameInput struct {
	focusIndex int
	inputs     []textinput.Model
	prevModel  tea.Model
}

func newPlayerNameInputModel(prevModel model) playerNameInput {
	m := playerNameInput{
		inputs:    make([]textinput.Model, prevModel.numberOfPlayers),
		prevModel: prevModel,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 25

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

func (m playerNameInput) Init() tea.Cmd {
	return textinput.Blink
}

func (m playerNameInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

				tui.Run(playerNames)

				return m.prevModel, nil
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

func (m playerNameInput) View() string {
	var b strings.Builder
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	submit := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true).
		Align(lipgloss.Center).
		Render("Submit")
	if m.focusIndex == len(m.inputs) {
		submit = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center).
			Render("Submit")
	}
	fmt.Fprintf(&b, "\n\n%s", submit)
	return b.String()
}

func (m *playerNameInput) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
