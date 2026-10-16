package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expenses-bot/backend/internal"
	"expenses-bot/backend/internal/api"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPAddr string `env:"HTTP_ADDR" envDefault:":8080"`
	DBPath   string `env:"DB_PATH" envDefault:"./data/expenses.db"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func main() {
	settings, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := internal.NewLogger(settings.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := internal.OpenSQLite(ctx, settings.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repository := internal.NewRepository(db)
	expenseService := api.NewExpenseService(repository)
	handlers := api.Handlers{Expenses: api.NewExpenseHandler(expenseService)}
	server := &http.Server{
		Addr:              settings.HTTPAddr,
		Handler:           api.NewRouter(logger, handlers),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shut down HTTP server", "error", err)
		}
	}()

	slog.Info("starting api", "addr", settings.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to serve API", "error", err)
		os.Exit(1)
	}
}
