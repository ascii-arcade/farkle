package menu

import "errors"

type MenuError error

var (
	ErrInvalidCode MenuError = errors.New("invalid_code")
	ErrLobbyFull   MenuError = errors.New("lobby_full")
)
