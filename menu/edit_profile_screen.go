package menu

import (
	"strings"

	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type editProfileScreen struct {
	model *Model
	style lipgloss.Style

	userNameInput textinput.Model
	sshKeyInput   textarea.Model
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

	sshKeyInput := textarea.New()
	sshKeyInput.Cursor.Style = m.style.Foreground(lipgloss.Color("205"))
	sshKeyInput.Placeholder = ""
	sshKeyInput.CharLimit = 600
	sshKeyInput.SetWidth(70)
	sshKeyInput.SetHeight(10)
	sshKeyInput.SetValue(m.player.SshPubKey)
	sshKeyInput.Blur()
	s.sshKeyInput = sshKeyInput

	return s
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
			switch s.cursorPos {
			case 0:
				s.userNameInput.Blur()
				s.sshKeyInput.Focus()
				s.cursorPos = 1
			case 1:
				s.sshKeyInput.Blur()
				s.userNameInput.Focus()
				s.cursorPos = 0
			}
		}

		if keys.PreviousScreen.TriggeredBy(msg.String()) {
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newOptionScreen(),
				}
			}
		}

		s.model.player.AddPubKey(s.sshKeyInput.Value())
	}

	s.sshKeyInput, cmd = s.sshKeyInput.Update(msg)

	s.sshKeyInvalid = false
	if s.sshKeyInput.Value() != "" && !utils.ValidPublicKey(s.sshKeyInput.Value()) {
		s.sshKeyInvalid = true
	}

	return s.model, cmd
}

func (s *editProfileScreen) View() string {
	var content strings.Builder
	content.WriteString("Set username | ")
	content.WriteString(s.userNameInput.View())
	content.WriteString("\n\n")

	content.WriteString("Set SSH Key\n")
	content.WriteString(s.sshKeyInput.View())
	if s.sshKeyInvalid {
		content.WriteString(" (invalid)")
	}
	content.WriteString("\n\n")

	return content.String()
}
