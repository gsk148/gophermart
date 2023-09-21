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

	h.Router.Group(func(r chi.Router) {
		r.Use(auth.Authorization)
		r.Post("/api/user/orders", h.LoadOrders)

		r.Get("/api/user/orders", h.GetOrders)
	})
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

func (h *Handler) LoadOrders(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromToken(w, r)
	contentType := r.Header["Content-Type"]
	if contentType[0] != "text/plain" {
		h.logger.Warn("Received non text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("PostURL: error: %s while reading body", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNum := string(body)
	h.logger.Info("LoadOrders: order num in body", orderNum)

	// TODO add validation for orderNum

	existingOrder, err := h.store.GetOrderByNumber(orderNum)
	if err != nil {
		switch err {
		case storage.ErrNoDBResult:
			{
				break
			}
		default:
			{
				h.logger.Error(fmt.Sprintf("LoadOrder: failed to check if exist, %s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	if existingOrder != nil && existingOrder.UserID == userID {
		h.logger.Warn(fmt.Sprintf("Provided order num %s already loaded with by user", orderNum))
		w.WriteHeader(200)
		return
	}
	if existingOrder != nil && existingOrder.UserID != userID {
		h.logger.Warn(fmt.Sprintf("Provided order num %s already exist", orderNum))
		w.WriteHeader(http.StatusConflict)
		return
	}

	// загрузили заказ
	if err = h.store.LoadOrder(orderNum, userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info(fmt.Sprintf("LoadOrders: saved order with number", orderNum))
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	userID := auth.GetUserIDFromToken(w, r)

	orders, err := h.store.GetOrdersByUserID(userID)
	if err != nil {
		switch err {
		case storage.ErrNoDBResult:
			{
				h.logger.Info(fmt.Sprintf("GetOrdersByUserID: no orders for user with id %s", userID))
				w.WriteHeader(http.StatusNoContent)
				return
			}
		default:
			{
				h.logger.Error(fmt.Sprintf("getUserInfoByToken: failed, %s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	resp, err := json.Marshal(orders)
	if err != nil {
		h.logger.Error("GetOrders: failed to marshal resp %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
