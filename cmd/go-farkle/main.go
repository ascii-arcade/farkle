package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"

	"github.com/ascii-arcade/farkle/internal/menu"
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

	logger.Info(serverUrl)
}

func main() {
	playerNames := flag.Args()
	if len(playerNames) > 0 {
		rand.Shuffle(len(playerNames), func(i, j int) {
			playerNames[i], playerNames[j] = playerNames[j], playerNames[i]
		})
		splashScreen.Run()
		tui.Run(playerNames)
		return
	}

	splashScreen.Run()
	menu.Run()
}
