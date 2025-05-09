package server

import "net/http"

func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Health check received")

	w.Write([]byte("OK"))
}
