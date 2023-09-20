package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"

	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/models"
	"github.com/gsk148/gophermart/internal/storage"
	"github.com/gsk148/gophermart/internal/utils"
)

type Handler struct {
	Router *chi.Mux
	store  *storage.Storage
	logger *slog.Logger
}

func NewHandler(router *chi.Mux, store *storage.Storage, logger *slog.Logger) *Handler {
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
	h.Router.Post("/api/users/login", h.login)
	h.Router.Post("/api/users/register", h.register)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	body, err := io.ReadAll(r.Body)

	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read from body %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to unmarshal body %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userDB, err := h.store.GetUserByLogin(user.Login)
	if !utils.CheckHashAndPassword(userDB.Password, user.Password) {
		h.logger.Error(fmt.Sprintf("Failed to authorization %s", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err != nil && err != storage.ErrNoDBResult {
		h.logger.Error("Failed to login, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = auth.GenerateCookie(w, user.ID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to generate cookie, %s", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Register: failed to read from body %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Register: failed to unmarshal body %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userInDB, err := h.store.GetUserByLogin(user.Login)
	switch {
	case err == nil && userInDB.Login != "":
		h.logger.Warn(fmt.Sprintf("Register: user with provided login %s exists", user.Login))
		w.WriteHeader(http.StatusConflict)
		return
	case err == storage.ErrNoDBResult:
		cryptedPsw, err := utils.HashString(user.Password)
		if err != nil {
			h.logger.Error("Register: failed to encrypt password")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID, err := h.store.Register(user.Login, cryptedPsw)
		if err != nil {
			h.logger.Error("Register: failed while registering in storage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = auth.GenerateCookie(w, userID)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to generate cookie, %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
