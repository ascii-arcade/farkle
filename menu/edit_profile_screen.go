package menu

import (
	"strings"

	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/players"
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

	userNameInput   textinput.Model
	sshKeyNameInput textinput.Model
	sshKeyInput     textinput.Model
	cursorPos       int

	manageKeys   bool
	sshKeysTable table.Model

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

	sshKeyNameInput := textinput.New()
	sshKeyNameInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	sshKeyNameInput.CharLimit = 20
	sshKeyNameInput.Blur()
	s.sshKeyNameInput = sshKeyNameInput

	sshKeyInput := textinput.New()
	sshKeyInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	sshKeyInput.Placeholder = ""
	sshKeyInput.CharLimit = 600
	sshKeyInput.Width = 50
	sshKeyInput.Blur()
	s.sshKeyInput = sshKeyInput

	columns := []table.Column{
		{Title: "", Width: 10},
		{Title: "Public SSH Keys", Width: 40},
	}

	rows := []table.Row{}

	for name, key := range m.player.SshPubKeys {
		rows = append(rows, table.Row{name, key})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = s.style.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = s.style.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(tableStyle)

	s.sshKeysTable = t

	return s
}

func (s *editProfileScreen) refreshSSHKeysTable() {
	rows := make([]table.Row, len(s.model.player.SshPubKeys))

	for name, key := range s.model.player.SshPubKeys {
		rows = append(rows, table.Row{name, key})
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

			if s.manageKeys {
				s.sshKeysTable, cmd = s.sshKeysTable.Update(msg)
				return s.model, cmd
			}

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
				s.sshKeyNameInput.Blur()
				s.sshKeyInput.Blur()
			case 1:
				s.userNameInput.Blur()
				s.sshKeyNameInput.Focus()
				s.sshKeyInput.Blur()
			case 2:
				s.userNameInput.Blur()
				s.sshKeyNameInput.Blur()
				s.sshKeyInput.Focus()
			}
		case "m":
			s.userNameInput.Blur()
			s.sshKeyNameInput.Blur()
			s.sshKeyInput.Blur()
			s.sshKeysTable.Focus()
			s.manageKeys = true
			return s.model, nil
		case "enter":
			switch s.cursorPos {
			case 0:
				s.userNameInput, cmd = s.userNameInput.Update(msg)
				s.model.player.Username = s.userNameInput.Value()
				_ = s.model.player.Save()
				return s.model, cmd
			case 1, 2:
				if s.sshKeyNameInput.Value() == "" || s.sshKeyInput.Value() == "" {
					s.model.error = "Both SSH Key Name and Key are required"
					return s.model, nil
				}

				s.sshKeyInput, cmd = s.sshKeyInput.Update(msg)

				if !utils.ValidPublicKey(s.sshKeyInput.Value()) {
					return s.model, cmd
				}

				if _, exists := s.model.player.SshPubKeys[s.sshKeyNameInput.Value()]; exists {
					s.model.error = "SSH Key Name already exists"
					return s.model, nil
				}

				key := strings.Join(strings.Split(s.sshKeyInput.Value(), " ")[:2], " ")
				if _, found := players.Get(key); found {
					s.model.error = "SSH Key already in use"
					return s.model, nil
				}

				s.model.player.SshPubKeys[s.sshKeyNameInput.Value()] = s.sshKeyInput.Value()
				_ = s.model.player.Save()

				s.sshKeyNameInput.SetValue("")
				s.sshKeyNameInput.Blur()
				s.sshKeyInput.SetValue("")
				s.sshKeyInput.Blur()
				s.userNameInput.Focus()
				s.cursorPos = 0

				s.refreshSSHKeysTable()
				return s.model, cmd
			}
		case "c":
			s.sshKeyNameInput.SetValue("")
			s.sshKeyNameInput.Blur()
			s.sshKeyInput.SetValue("")
			s.sshKeyInput.Blur()
			s.userNameInput.Focus()
			s.cursorPos = 0

		case "d":
			if s.manageKeys {
				if len(s.model.player.SshPubKeys) <= 1 {
					s.model.error = "You must have at least one Public SSH key"
					return s.model, nil
				}

				keyName := s.sshKeysTable.SelectedRow()[0]
				delete(s.model.player.SshPubKeys, keyName)
				_ = s.model.player.Save()
				s.refreshSSHKeysTable()
				s.sshKeysTable, cmd = s.sshKeysTable.Update(msg)
				return s.model, cmd
			}
		}

		if keys.PreviousScreen.TriggeredBy(msg.String()) {
			if s.manageKeys {
				s.manageKeys = false
				return s.model, nil
			}

			_ = s.model.player.Save()

			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newOptionScreen(),
				}
			}
		}
	}

	switch s.cursorPos {
	case 0:
		s.userNameInput, cmd = s.userNameInput.Update(msg)
		s.model.player.Username = s.userNameInput.Value()
	case 1:
		s.sshKeyNameInput, cmd = s.sshKeyNameInput.Update(msg)
	case 2:
		s.sshKeyInput, cmd = s.sshKeyInput.Update(msg)

		s.sshKeyInvalid = false
		if s.sshKeyInput.Value() != "" && !utils.ValidPublicKey(s.sshKeyInput.Value()) {
			s.sshKeyInvalid = true
		}
	}

	s.model.error = ""
	return s.model, cmd
}

func (s *editProfileScreen) View() string {
	if s.manageKeys {
		var content strings.Builder
		content.WriteString(s.sshKeysTable.View())
		if s.model.error != "" {
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(s.model.error))
			s.model.error = ""
		}
		content.WriteString("\n\nPress 'esc' to return to profile editing")

		style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))

		return style.Render(content.String())
	}

	var content strings.Builder
	content.WriteString("Set username | ")
	content.WriteString(s.userNameInput.View())
	content.WriteString("\n\n")

	content.WriteString("Add SSH Key\n")
	content.WriteString("   Name | ")
	content.WriteString(s.sshKeyNameInput.View())
	content.WriteString("\n   Key  | ")
	content.WriteString(s.sshKeyInput.View())
	if s.sshKeyInvalid {
		content.WriteString(" (invalid)")
	}
	if s.model.error != "" {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(s.model.error))
	}
	content.WriteString("\n\n'c' to clear inputs | 'm' to manage SSH keys")

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))

	return style.Render(content.String())
}
