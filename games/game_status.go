package games

type GameStatus string

const (
	StatusWaitingForPlayers GameStatus = "waiting_for_players"
	StatusInProgress        GameStatus = "in_progress"
	StatusCompleted         GameStatus = "completed"
)

func (gs GameStatus) String() string {
	switch gs {
	case StatusWaitingForPlayers:
		return "Waiting for Players"
	case StatusInProgress:
		return "In Progress"
	case StatusCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}
