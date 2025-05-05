package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/websocket"
)

var (
	debug = false

	logger *slog.Logger
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()

	logHandlerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	if debug {
		logHandlerOpts.Level = slog.LevelDebug
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &logHandlerOpts))
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s := websocket.Server{
			Handler: websocket.Handler(func(ws *websocket.Conn) {
				logger.Info("WebSocket connection established", "remoteAddr", ws.RemoteAddr())
				defer ws.Close()

				for {
					var msg string
					if err := websocket.Message.Receive(ws, &msg); err != nil {
						logger.Error("Failed to receive message", "error", err)
						break
					}
					logger.Info("Received message", "message", msg)

					if err := websocket.Message.Send(ws, msg); err != nil {
						logger.Error("Failed to send message", "error", err)
						break
					}
				}
			})}
		s.ServeHTTP(w, r)
	})

	go func() {
		logger.Info("Starting server", "debug", debug)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("Failed to start server", "error", err)
		}
	}()

	// watch for ctl-c
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	logger.Info("Shutting down server")
}
