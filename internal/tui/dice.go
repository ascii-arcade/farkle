package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var dieCharacters = map[int]string{
	1: "⚀",
	2: "⚁",
	3: "⚂",
	4: "⚃",
	5: "⚄",
	6: "⚅",
}

func left(s string) string {
	return lipgloss.NewStyle().Width(5).Align(lipgloss.Left).Render(s)
}

func center(s string) string {
	return lipgloss.NewStyle().Width(5).Align(lipgloss.Center).Render(s)
}

func right(s string) string {
	return lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render(s)
}

func die(face int) string {
	return lipgloss.NewStyle().
		Width(7).
		Height(3).
		Align(lipgloss.Center).
		Border(lipgloss.Border(lipgloss.RoundedBorder())).
		MarginRight(1).
		Render(dieFaces[face])
}

var dieFaces = map[int]string{
	1: "\n" + center("●"),
	2: strings.Join([]string{left("●"), "", right("●")}, "\n"),
	3: strings.Join([]string{left("●"), center("●"), right("●")}, "\n"),
	4: strings.Join([]string{center("●   ●"), "", center("●   ●")}, "\n"),
	5: strings.Join([]string{center("●   ●"), center("●"), center("●   ●")}, "\n"),
	6: strings.Join([]string{center("●   ●"), center("●   ●"), center("●   ●")}, "\n"),
}
