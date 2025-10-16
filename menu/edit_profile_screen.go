package menu

import (
	"fmt"
	"strings"

	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type editProfileScreen struct {
	model *Model
	style lipgloss.Style

	userNameInput textinput.Model
	sshKeyInput   textinput.Model
	sshKeysTable  table.Model
	cursorPos     int

	sshKeyInvalid bool
}

func (m *Model) newEditProfileScreen() *editProfileScreen {
	s := &editProfileScreen{
		model: m,
		style: m.style,
	}

	userNameInput := textinput.New()
	userNameInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	userNameInput.Placeholder = ""
	userNameInput.CharLimit = 20
	userNameInput.Focus()
	userNameInput.SetValue(m.player.Username)
	s.userNameInput = userNameInput

	sshKeyInput := textinput.New()
	sshKeyInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	sshKeyInput.Placeholder = ""
	sshKeyInput.CharLimit = 600
	sshKeyInput.Blur()
	s.sshKeyInput = sshKeyInput

	columns := []table.Column{
		{Title: "", Width: 2},
		{Title: "Public SSH Keys", Width: 40},
	}

	rows := make([]table.Row, len(s.model.player.SshPubKeys))

	for i, key := range s.model.player.SshPubKeys {
		rows[i] = table.Row{fmt.Sprintf("%d", i+1), key}
	}

	if len(rows) == 0 {
		rows = append(rows, table.Row{"-", "No SSH Keys Added"})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(5),
		table.WithFocused(false),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	tableStyle.Selected = tableStyle.Selected.
		Bold(false).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("7"))
	t.SetStyles(tableStyle)

	s.sshKeysTable = t

	return s
}

func (s *editProfileScreen) refreshSSHKeysTable() {
	rows := make([]table.Row, len(s.model.player.SshPubKeys))

	for i, key := range s.model.player.SshPubKeys {
		rows[i] = table.Row{fmt.Sprintf("%d", i+1), key}
	}

	if len(rows) == 0 {
		rows = append(rows, table.Row{"-", "No SSH Keys Added"})
	}

	s.sshKeysTable.SetRows(rows)
}

func (s *editProfileScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *editProfileScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case cursor.BlinkMsg:
		return s.model, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				s.cursorPos--
			} else {
				s.cursorPos++
			}

			if s.cursorPos > 2 {
				s.cursorPos = 0
			} else if s.cursorPos < 0 {
				s.cursorPos = 2
			}

			switch s.cursorPos {
			case 0:
				s.userNameInput.Focus()
				s.sshKeyInput.Blur()
				s.sshKeysTable.Blur()
			case 1:
				s.userNameInput.Blur()
				s.sshKeyInput.Focus()
				s.sshKeysTable.Blur()
			case 2:
				s.userNameInput.Blur()
				s.sshKeyInput.Blur()
				s.sshKeysTable.Focus()
				s.sshKeysTable.SetCursor(0)
			}
			return s.model, nil
		}

		if keys.PreviousScreen.TriggeredBy(msg.String()) {
			_ = s.model.player.Save()

			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newOptionScreen(),
				}
			}
		}

		s.model.player.AddPubKey(s.sshKeyInput.Value())
	}

	switch s.cursorPos {
	case 0:
		s.userNameInput, cmd = s.userNameInput.Update(msg)
		s.model.player.Username = s.userNameInput.Value()
		s.model.player.Save()
	case 1:
		s.sshKeyInput, cmd = s.sshKeyInput.Update(msg)

		s.sshKeyInvalid = false
		if s.sshKeyInput.Value() != "" && !utils.ValidPublicKey(s.sshKeyInput.Value()) {
			s.sshKeyInvalid = true
		}
	case 2:
		s.refreshSSHKeysTable()
		s.sshKeysTable, cmd = s.sshKeysTable.Update(msg)
	}

	return s.model, cmd
}

func (s *editProfileScreen) View() string {
	var content strings.Builder
	content.WriteString("Set username | ")
	content.WriteString(s.userNameInput.View())
	content.WriteString("\n\n")

	content.WriteString("Add SSH Key  | ")
	content.WriteString(s.sshKeyInput.View())
	if s.sshKeyInvalid {
		content.WriteString(" (invalid)")
	}
	content.WriteString("\n\n")

	content.WriteString(s.sshKeysTable.View())

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))

	return style.Render(content.String())
}
