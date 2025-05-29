package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ascii-arcade/farkle/root"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	tea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

var (
	debug   = false
	logCli  = false
	logFile = "log"
	logPath = "/var/log/ascii-arcade"
	host    = "localhost"
	port    = "23234"
	logger  *slog.Logger
)

func init() {
	// Set up configuration
	debugStr := os.Getenv("ASCII_ARCADE_DEBUG")
	if debugStr != "" {
		if debugStr == "true" || debugStr == "1" {
			debug = true
		}
	}

	logCliStr := os.Getenv("ASCII_ARCADE_LOG_CLI")
	if logCliStr != "" {
		if logCliStr == "true" || logCliStr == "1" {
			logCli = true
		}
	}
	logFile = os.Getenv("ASCII_ARCADE_LOG_FILE")
	logPath = os.Getenv("ASCII_ARCADE_LOG_PATH")

	host = os.Getenv("ASCII_ARCADE_HOST")
	port = os.Getenv("ASCII_ARCADE_PORT")

	slogLevel := slog.LevelInfo
	if debug {
		slogLevel = slog.LevelDebug
	}

	// Set up logging
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})
	if debug {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	}

	logger = slog.New(handler)
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			tea.Middleware(root.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		logger.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("Starting SSH server", "host", host, "port", port)

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	logger.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		logger.Error("Could not stop server", "error", err)
	}
}
