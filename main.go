package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"libdb.so/hserve"
	"libdb.so/scouts-server/api"
	"libdb.so/scouts-server/api/gameserver"
	"libdb.so/scouts-server/api/storage"
	"libdb.so/scouts-server/api/user"
)

var (
	httpAddr = "localhost:8080"
	stateDir = "/tmp/scouts-server"
	verbose  = false
)

func init() {
	flag.StringVar(&httpAddr, "http", httpAddr, "HTTP address to listen on")
	flag.StringVar(&stateDir, "state", stateDir, "state directory")
	flag.BoolVar(&verbose, "verbose", verbose, "verbose logging")
	flag.Parse()
}

func main() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:   level,
		NoColor: !isatty.IsTerminal(os.Stderr.Fd()),
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := start(ctx, logger); err != nil {
		logger.Error("failed to start", "error", err)
		os.Exit(1)
	}
}

func start(ctx context.Context, logger *slog.Logger) error {
	storageManager := storage.NewStorageManager(stateDir)

	sessionStorage, err := storageManager.OpenSessionStorage()
	if err != nil {
		return fmt.Errorf("failed to open session storage: %w", err)
	}

	gameManager := gameserver.NewGameManager(logger)

	api := api.NewHandler(api.Services{
		GameManager:    gameManager,
		SessionStorage: user.NewCachedSessionStorage(sessionStorage),
	})

	r := chi.NewMux()
	r.Mount("/api/v1", api)

	if err := hserve.ListenAndServe(ctx, httpAddr, r); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}
