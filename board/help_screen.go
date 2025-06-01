package board

import (
	"github.com/ascii-arcade/farkle/dice"
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
	objective := "Be the player with the highest score."

	scoring := dice.GetDieCharacter(1) + " = 100 points\n" +
		dice.GetDieCharacter(5) + " = 50 points\n" +
		dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " " + dice.GetDieCharacter(1) + " = 300 points\n" +
		dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " " + dice.GetDieCharacter(2) + " = 200 points\n" +
		dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " " + dice.GetDieCharacter(3) + " = 300 points\n" +
		dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " " + dice.GetDieCharacter(4) + " = 400 points\n" +
		dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " " + dice.GetDieCharacter(5) + " = 500 points\n" +
		dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " " + dice.GetDieCharacter(6) + " = 600 points\n" +
		"Four of a kind = 1000 points\n" +
		"Five of a kind = 2000 points\n" +
		"Six of a kind = 3000 points\n" +
		"Straight (1-6) = 1500 points\n" +
		"Three pairs = 1500 points\n" +
		"Four of a kind + pair = 1500 points\n" +
		"Two triplets = 2500 points\n"

	content := "Help\n"
	content += "Press 'q' to return to the game.\n\n"
	content += "Objective:\n" + objective + "\n\n"
	content += "Scoring:\n" + scoring

	return s.model.style.Render(
		lipgloss.Place(
			s.model.width,
			s.model.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		),
	)

}

func (s *helpScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	return s.model, nil
}
