package server

import (
	"net/http"
	"time"

	"github.com/ascii-arcade/farkle/internal/player"
	"golang.org/x/net/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			logger.Info("webSocket connection established")
			defer ws.Close()

			name := r.URL.Query().Get("name")
			player := player.NewPlayer(ws, name)
			logger.Info("new client connected", "clientId", player.Id)

			h.register <- player

			for player.Active {
				time.Sleep(1 * time.Second)
			}
		})}
	s.ServeHTTP(w, r)

}
