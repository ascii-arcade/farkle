package server

import (
	"log/slog"
	"net/http"
)

var logger *slog.Logger
var h *hub
var lobbies map[string]*lobby

func Run(l *slog.Logger, debug bool) {
	logger = l

	h = newHub(logger)
	go h.monitorBroadcast()
	go h.monitorConnections()
	go h.run()

	lobbies = make(map[string]*lobby)

	http.HandleFunc("/ws", wsHandler)

	logger.Info("Starting server", "debug", debug)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}
