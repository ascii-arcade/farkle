package keys

import (
	"slices"

	"github.com/charmbracelet/lipgloss"
)

type Keys []string

func (k Keys) TriggeredBy(msg string) bool {
	return slices.Contains(k, msg)
}

func (k Keys) String(style lipgloss.Style) string {
	if len(k) == 0 {
		return ""
	}
	return style.Bold(true).Italic(true).Render(k[0])
}

var (
	ExitApplication = Keys{"ctrl+c"}

	MenuJoinGame     = Keys{"j"}
	MenuStartNewGame = Keys{"n"}
	MenuEnglish      = Keys{"1"}
	MenuSpanish      = Keys{"2"}

	Back = Keys{"q"}

	PreviousScreen = Keys{"esc"}
	Submit         = Keys{"enter"}

	LobbyStartGame = Keys{"s"}
	RestartGame    = Keys{"r"}

	ActionLock = Keys{"l"}
	ActionBank = Keys{"b", "y"}
	ActionRoll = Keys{"r"}
	ActionUndo = Keys{"u", "backspace"}
	OpenHelp   = Keys{"?"}
)
