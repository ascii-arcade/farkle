package server

import (
	"log/slog"
	"net/http"
)

var logger *slog.Logger
var h *hub

func Run(l *slog.Logger, debug bool) {
	logger = l

	h = newHub(logger)
	go h.monitorConnections()
	go h.monitorBroadcast()
	go h.monitorLobbies()
	go h.run()

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/health", healthHandler)

	logger.Info("Starting server", "debug", debug)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}
