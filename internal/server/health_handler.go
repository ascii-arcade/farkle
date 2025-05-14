package server

import "net/http"

func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	logger.Debug("Health check received")

	_, _ = w.Write([]byte("OK"))
}
