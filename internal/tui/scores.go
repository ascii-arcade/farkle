package tui

import (
	"strconv"
	"strings"
)

func (m *model) playerScores() string {
	scores := make([]string, len(m.players))

	for i, player := range m.players {
		if i == m.currentPlayerIndex {
			scores[i] = player.name + ": " + strconv.Itoa(player.score) + " (current)"
		} else {
			scores[i] = player.name + ": " + strconv.Itoa(player.score)
		}
	}

	return strings.Join(scores, "\n\n")
}
