package lobbies

import "strings"

func GetLobby(code string) *Lobby {
	lobby, exists := lobbies[strings.ToUpper(code)]
	if !exists {
		return nil
	}
	return lobby
}
