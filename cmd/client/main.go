package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"

	"github.com/ascii-arcade/farkle/client"
	"github.com/ascii-arcade/farkle/client/menu"
	"github.com/ascii-arcade/farkle/config"
	splashScreen "github.com/ascii-arcade/farkle/splash_screen"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	logger *slog.Logger
)

func init() {
	serverURL := flag.String("server", "localhost", "WebSocket server URL")
	serverPort := flag.String("port", "8080", "WebSocket server port")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	config.SetServerURL(serverURL)
	config.SetServerPort(serverPort)
	config.SetDebug(debug)

	loggerHandler := slog.NewTextHandler(io.Discard, nil)
	if *debug {
		syslogWriter, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL7, "farkle")
		if err != nil {
			fmt.Println("Error opening syslog:", err)
			goto CONTINUE
		}

		logOpts := slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}

		loggerHandler = slog.NewTextHandler(syslogWriter, &logOpts)
	}
CONTINUE:

	logger = slog.New(loggerHandler)
}

func main() {
	initModel := client.App{
		CurrentView: menu.New(),
	}
	p := tea.NewProgram(initModel)

	splashScreen.Run()
	if _, err := p.Run(); err != nil {
		logger.Error("Error running client", "error", err)
	}
}
