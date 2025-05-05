package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ascii-arcade/farkle/internal/server"
)

var (
	debug = false
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()
}

func main() {
	logHandlerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	if debug {
		logHandlerOpts.Level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &logHandlerOpts))
	go server.Run(logger, debug)

	// watch for ctl-c
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	logger.Info("Shutting down server")
}
