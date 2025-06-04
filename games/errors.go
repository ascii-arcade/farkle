package games

import "errors"

type GameError error

var (
	ErrGameNotFound          GameError = errors.New("not_found")
	ErrGameAlreadyInProgress GameError = errors.New("already_in_progress")
	ErrScoreTooLow           GameError = errors.New("score_too_low")
	ErrNoDiceHeld            GameError = errors.New("no_dice_held")
)
