package server

import (
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			logger.Info("webSocket connection established", "remoteAddr", ws.RemoteAddr())
			defer ws.Close()

			client := h.newClient(ws)
			logger.Info("new client connected", "clientId", client.id)

			for client.active {
				time.Sleep(1 * time.Second)
			}
		})}
	s.ServeHTTP(w, r)

}
