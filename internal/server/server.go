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
	go h.monitorBroadcast()
	go h.monitorConnections()
	go h.startTimedBroadcast()
	go h.run()

	http.HandleFunc("/ws", wsHandler)

	logger.Info("Starting server", "debug", debug)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}
