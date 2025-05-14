package lobbies

func GetLobby(code string) *Lobby {
	lobby, exists := lobbies[code]
	if !exists {
		return nil
	}
	return lobby
}
