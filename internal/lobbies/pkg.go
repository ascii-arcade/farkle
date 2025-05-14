package lobbies

import (
	"math/rand"
	"sync"
)

var lobbies = make(map[string]*Lobby)
var mu = &sync.Mutex{}

func AddLobby(l *Lobby) {
	lobbies[l.Code] = l
}

func newCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b[:3]) + "-" + string(b[3:6])
}
