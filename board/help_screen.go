package board

import (
	"strings"

	"github.com/ascii-arcade/farkle/dice"
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
	objective.WriteString("Objective\n")
	objective.WriteString("First to 10,000 points triggers the end of the game.\n")
	objective.WriteString("After end game starts, rest of players get one last turn.\n")
	objective.WriteString("The player with the most points at the end of the game wins.\n\n")

	rules := strings.Builder{}
	rules.WriteString("Rules\n")
	rules.WriteString("Players take turns rolling six dice.\n")
	rules.WriteString("On each turn, the player rolls and choosing any dice that can be scored (holding).\n")
	rules.WriteString("The player can hold any number of dice, but must hold at least one die.\n")
	rules.WriteString("The player can then keep the dice (locking).\n")
	rules.WriteString("The player then either rolls the remaining dice or stops and score the points (banking).\n")
	rules.WriteString("The first time a player banks points, it must equal 500 or more.\n")
	rules.WriteString("If the player rolls and can not score any points, they lose all points for that turn (bust).\n\n")

	scoring := []string{dice.GetDieCharacter(1) + " = 100 points",
		dice.GetDieCharacter(5) + " = 50 points",
		dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " = 300 points",
		dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " = 200 points",
		dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " = 300 points",
		dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " = 400 points",
		dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " = 500 points",
		dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " = 600 points",
		"Four of a kind = 1000 points",
		"Five of a kind = 2000 points",
		"Six of a kind = 3000 points",
		"Straight (1-6) = 1500 points",
		"Three pairs = 1500 points",
		"Four of a kind + pair = 1500 points",
		"Two triplets = 2500 points",
	}

	content := strings.Builder{}
	content.WriteString("Help\n")
	content.WriteString("Press 'q' to return to the game.\n\n")
	content.WriteString(objective.String())
	content.WriteString(rules.String())
	content.WriteString("Scoring\n" + s.model.style.Render(
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
