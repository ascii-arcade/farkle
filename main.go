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

	"github.com/ascii-arcade/farkle/app"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/web"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	tea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

var (
	logger *slog.Logger
)

func init() {
	config.Setup()

	slogLevel := slog.LevelInfo
	if config.GetDebug() {
		slogLevel = slog.LevelDebug
	}

	// Set up logging
	handlerOpts := &slog.HandlerOptions{
		Level: slogLevel,
	}
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	if config.GetDebug() {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	logger = slog.New(handler)
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(config.GetServerHost(), config.GetServerPortSSH())),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			tea.Middleware(app.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		logger.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("Starting SSH server", "host", config.GetServerHost(), "port", config.GetServerPortSSH())

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	go web.Run()

	<-done
	logger.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		logger.Error("Could not stop server", "error", err)
	}
}
