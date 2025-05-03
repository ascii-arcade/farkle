package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func logPaneStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(80).
		Height(15)
}

func (l *log) entries() string {
	if len(*l) <= 15 {
		return strings.Join(*l, "\n")
	}

	recent := (*l)[len(*l)-15:]

	return strings.Join(recent, "\n")
}

func (l *log) add(entry string) {
	*l = append(*l, entry)
}

func (m *model) logPane() string {
	return logPaneStyle().Render(m.log.entries())
}
