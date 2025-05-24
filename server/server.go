package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func Run(logger *slog.Logger, debug bool) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws/{lobbyId}", wsHandler)
	mux.HandleFunc("/lobbies", lobbiesHandler)
	mux.HandleFunc("/health", healthHandler)

	loggedMux := loggerMiddleware(logger)(mux)

	logger.Info("Starting server", "debug", debug)
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}

func loggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := context.WithValue(r.Context(), "LOGGER", logger)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info("Request handled", "method", r.Method, "path", r.URL.Path, "duration", duration)
		})
	}
}

func getLogger(r *http.Request) *slog.Logger {
	logger, ok := r.Context().Value("LOGGER").(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
