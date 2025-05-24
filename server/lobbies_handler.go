package server

import (
	"encoding/json"
	"net/http"

	"github.com/ascii-arcade/farkle/lobbies"
)

func lobbiesHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lobby := lobbies.NewLobby(logger)
	lobbies.AddLobby(lobby)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(lobby); err != nil {
		http.Error(w, "Failed to encode lobby", http.StatusInternalServerError)
		return
	}
}
