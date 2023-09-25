package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gsk148/gophermart/internal/client"
	"golang.org/x/exp/slog"

	"github.com/gsk148/gophermart/internal/config"
	"github.com/gsk148/gophermart/internal/handlers"
	"github.com/gsk148/gophermart/internal/storage"
)

func main() {
	cfg := config.MustLoad()

	// add logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	db, err := storage.InitStorage(cfg.DatabaseAddress, *logger)
	if err != nil {
		logger.Error("failed to init storage")
		os.Exit(1)
	}

	router := chi.NewRouter()

	repository := storage.NewRepository(context.Background(), db, *logger)
	accrualClient := client.NewAccrualClient(repository, cfg.AccrualSystemAddress, 10)
	go accrualClient.Run()

	//add handler
	h := handlers.NewHandler(router, repository, logger)

	if err := http.ListenAndServe(cfg.RunAddress, h.Router); err != nil {
		logger.Error("failed to start server")
	}
}
