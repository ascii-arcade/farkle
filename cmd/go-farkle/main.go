package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"

	splashScreen "github.com/ascii-arcade/farkle/internal/splash_screen"
	"github.com/ascii-arcade/farkle/internal/tui"
)

var (
	serverUrl = "ws://localhost:8080/ws"

	logger *slog.Logger
)

func init() {
	flag.StringVar(&serverUrl, "server", serverUrl, "WebSocket server URL")
	flag.Parse()

	logFile, err := os.OpenFile("farkle.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}

	logOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	loggerHandler := slog.NewTextHandler(logFile, &logOpts)
	logger = slog.New(loggerHandler)
}

func main() {
	// c, err := wsclient.NewWsClient(serverUrl)
	// if err != nil {
	// 	logger.Error("Error connecting to server", "error", err)
	// 	return
	// }
	// defer c.Close()

	// go func() {
	// 	for {
	// 		if err := c.MonitorMessages(); err != nil {
	// 			logger.Error("Error monitoring messages", "error", err)

	// 			time.Sleep(5 * time.Second)
	// 			logger.Debug("Attempting to reconnect...")

	// 			if err := c.Reconnect(serverUrl); err != nil {
	// 				logger.Error("Error reconnecting", "error", err)
	// 			}
	// 		}
	// 	}
	// }()

	playerNames := os.Args[1:]
	if len(playerNames) == 0 {
		fmt.Println("Usage: go-farkle player1 player2 ...")
		return
	}

	rand.Shuffle(len(playerNames), func(i, j int) {
		playerNames[i], playerNames[j] = playerNames[j], playerNames[i]
	})
	splashScreen.Run()
	tui.Run(playerNames)
}
