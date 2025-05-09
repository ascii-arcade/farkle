package menu

import (
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type joinGameModel struct {
	width      int
	height     int
	focusIndex int

	inputs []textinput.Model
	errors string

	menuModel menuModel
	logger    *slog.Logger
	debug     bool
}

func newJoinGameModel(menuModel menuModel) joinGameModel {
	m := joinGameModel{
		width:      menuModel.width,
		height:     menuModel.height,
		focusIndex: 0,
		menuModel:  menuModel,
		logger:     menuModel.logger.With("component", "join_game"),
		inputs:     make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 25
		t.Width = 25

		switch i {
		case 0:
			t.Placeholder = "Your name"
			t.Focus()
		case 1:
			t.Placeholder = "Lobby code"
		}

		m.inputs[i] = t
	}

	return m
}

func (m joinGameModel) Init() tea.Cmd {
	return nil
}

func (m joinGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.menuModel, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tick(t)
			})
		case "tab", "down":
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = len(m.inputs)
			}
		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 0
			}
		case "enter":
			if m.focusIndex == len(m.inputs) {
				if m.inputs[0].Value() == "" {
					m.errors = "Please enter your name"
					m.inputs[0].Focus()
					m.focusIndex = 0
					return m, nil
				}
				if m.inputs[1].Value() == "" {
					m.errors = "Please enter a lobby code"
					m.inputs[1].Focus()
					m.focusIndex = 1
					return m, nil
				}
			}

			return m, nil
		}
	}

	cmd = m.updateInputs(msg)

	return m, cmd
}

func (m *joinGameModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	m.errors = ""

	return tea.Batch(cmds...)
}

func (m joinGameModel) View() string {
	paneStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-1).
		Align(lipgloss.Center, lipgloss.Center)

	if m.height < 15 || m.width < 100 {
		return paneStyle.Render("Window too small, please resize to something larger.")
	}

	inputsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1, 2)

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Align(lipgloss.Left, lipgloss.Bottom).
		Width(m.width)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff00")).
		Align(lipgloss.Left, lipgloss.Bottom)

	errorsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff0000")).
		Align(lipgloss.Left, lipgloss.Bottom)

	if m.debug {
		paneStyle = paneStyle.
			BorderForeground(lipgloss.Color("#ff0000")).
			BorderStyle(lipgloss.ASCIIBorder()).
			Height(m.height - 3).
			Width(m.width - 2)
	}

	inputs := []string{}
	for i := range m.inputs {
		if m.focusIndex == i {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
		inputs = append(inputs, m.inputs[i].View())
	}

	buttonPrefix := "   "
	if m.focusIndex == len(m.inputs) {
		buttonPrefix = "-> "
	}
	inputs = append(inputs, buttonStyle.Render(buttonPrefix+"Join Game"))

	inputPane := lipgloss.JoinVertical(
		lipgloss.Center,
		inputsStyle.Render(strings.Join(inputs, "\n")),
		errorsStyle.Render(m.errors),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		paneStyle.Render(inputPane),
		controlsStyle.Render("ESC to go back to menu"),
	)
}
