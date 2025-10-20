package main

import (
	"bytes"
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
	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/players"
	"github.com/ascii-arcade/farkle/web"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	tea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	gossh "golang.org/x/crypto/ssh"
)

func init() {
	config.Setup()

	slogLevel := slog.LevelInfo
	if config.GetDebug() {
		slogLevel = slog.LevelDebug
	}

	handlerOpts := &slog.HandlerOptions{
		Level: slogLevel,
	}
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	if config.GetDebug() {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	slog.SetDefault(slog.New(handler).With("app", "farkle", "version", config.Version))
}

func main() {
	ctx := context.Background()
	if err := database.Setup(ctx, config.GetDatabaseURI(), config.GetDatabase()); err != nil {
		slog.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	stopCleanup := make(chan struct{}, 1)
	defer close(stopCleanup)
	go games.StartCleanup(stopCleanup)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(config.GetServerHost(), config.GetServerPortSSH())),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			decodedKey := string(bytes.TrimSuffix(gossh.MarshalAuthorizedKey(key), []byte{'\n'}))
			slog.Debug("ssh authentication attempt", "user", ctx.User())

			player, found := players.Get(decodedKey)
			if !found {
				var err error
				if player, err = players.NewPlayer(ctx, "default", decodedKey, "en"); err != nil {
					slog.Error("could not create player", "error", err)
					return false
				}
				slog.Debug("created new player", "user", ctx.User())

				if ctx.User() == "web-client" {
					slog.Info("web-client ssh authentication successful", "user", ctx.User(), "player_id", player.Id)
					player.MakeVisitor()
				}
			}
			player.WithContext(ctx).Connect()

			ctx.SetValue("PLAYER", player)
			ctx.SetValue("PUBKEY", decodedKey)

			return true
		}),
		wish.WithMiddleware(
			tea.Middleware(app.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		slog.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	slog.Info("Starting SSH server", "host", config.GetServerHost(), "port", config.GetServerPortSSH())

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			slog.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	go func() {
		if err := web.Run(); err != nil {
			slog.Error("Could not start web server", "error", err)
			done <- nil
		}
	}()

	<-done
	slog.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		slog.Error("Could not stop server", "error", err)
	}
}
