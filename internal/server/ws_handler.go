package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobbies"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
	"golang.org/x/net/websocket"
)

var ErrLobbyFull error = errors.New("lobby is full")

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	lobbyId := r.PathValue("lobbyId")
	if lobbyId == "" {
		http.Error(w, "Lobby ID is required", http.StatusBadRequest)
		return
	}

	lobby := lobbies.GetLobby(lobbyId)
	if lobby == nil {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	s := websocket.Server{
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			logger.Info("webSocket connection established")
			defer ws.Close()

			player := player.NewPlayer(logger, ws, name)
			logger.Info("new client connected", "clientId", player.Id)

			if err := player.SendMessage(message.Message{
				Channel: message.ChannelPlayer,
				Type:    message.MessageTypeMe,
				Data:    player.ToJSON(),
				SentAt:  time.Now(),
			}); err != nil {
				logger.Error("Failed to send player message", "error", err)
			}

			if !lobby.AddPlayer(player) {
				if err := player.SendMessage(message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeError,
					Data:    ErrLobbyFull.Error(),
					SentAt:  time.Now(),
				}); err != nil {
					logger.Error("Failed to send error message", "error", err)
				}
				player.Close()
				return
			}

			lobby.BroadcastUpdate()

			for player.Active {
				time.Sleep(1 * time.Second)
			}
		})}
	s.ServeHTTP(w, r)

}
