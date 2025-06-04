package board

import (
	"fmt"
	"strings"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpScreen struct {
	model *Model
}

func (s *helpScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *helpScreen) View() string {
	objective := strings.Builder{}
	objective.WriteString(s.model.lang().Get("help", "objective", "title") + "\n")
	objective.WriteString(s.model.lang().Get("help", "objective", "text") + "\n\n")

	rules := strings.Builder{}
	rules.WriteString(s.model.lang().Get("help", "rules", "title") + "\n")
	rules.WriteString(s.model.lang().Get("help", "rules", "text") + "\n\n")

	scoring := []string{dice.GetDieCharacter(1) + " = 100 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(5) + " = 50 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " = 300 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " = 200 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " = 300 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " = 400 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " = 500 " + s.model.lang().Get("help", "scoring", "points"),
		dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " = 600 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "four_of_a_kind") + " = 1000 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "five_of_a_kind") + " = 2000 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "six_of_a_kind") + " = 3000 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "straight") + " = 1500 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "three_pairs") + " = 1500 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "four_of_a_kind_pair") + " = 1500 " + s.model.lang().Get("help", "scoring", "points"),
		s.model.lang().Get("help", "scoring", "two_triplets") + " = 2500 " + s.model.lang().Get("help", "scoring", "points"),
	}

	content := strings.Builder{}
	content.WriteString(s.model.lang().Get("help", "title") + "\n\n")
	content.WriteString(fmt.Sprintf(s.model.lang().Get("help", "back"), keys.Back) + "\n\n")
	content.WriteString(objective.String())
	content.WriteString(rules.String())
	content.WriteString(s.model.lang().Get("help", "scoring", "title") + "\n" + s.model.style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			scoring...,
		)),
	)

	return s.model.style.Render(
		lipgloss.Place(
			s.model.width,
			s.model.height,
			lipgloss.Center,
			lipgloss.Center,
			content.String(),
		),
	)

}

func (s *helpScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: &tableScreen{},
				}
			}
		}
	}
	return s.model, nil
}
