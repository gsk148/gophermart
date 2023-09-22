package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/labstack/gommon/log"
	"golang.org/x/exp/slog"

	"github.com/gsk148/gophermart/internal/config"
	"github.com/gsk148/gophermart/internal/handlers"
	"github.com/gsk148/gophermart/internal/storage"
)

func main() {
	cfg := config.MustLoad()

	db, err := storage.InitStorage(cfg.DatabaseAddress)
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}

	router := chi.NewRouter()
	// add logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	//add handler
	h := handlers.NewHandler(router, db, logger)

	if err := http.ListenAndServe(cfg.RunAddress, h.Router); err != nil {
		logger.Error("failed to start server")
	}
}
