package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"

	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/storage"
)

type Handler struct {
	Router     *chi.Mux
	repository *storage.Repository
	logger     *slog.Logger
}

func NewHandler(router *chi.Mux, store *storage.Repository, logger *slog.Logger) *Handler {
	h := &Handler{
		router,
		store,
		logger,
	}
	h.init()

	return h
}

func (h *Handler) init() {
	h.Router.Use(middleware.Logger)
	h.Router.Use(middleware.Compress(5, "text/html",
		"application/x-gzip",
		"text/plain",
		"application/json"))
	h.Router.Post("/api/user/register", h.Register)
	h.Router.Post("/api/user/login", h.Login)

	h.Router.Group(func(r chi.Router) {
		r.Use(auth.Authorization)
		r.Post("/api/user/orders", h.LoadOrders)
		r.Post("/api/user/balance/withdraw", h.DeductPoints)

		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalance)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})
}
