package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/errors"
	"github.com/gsk148/gophermart/internal/models"
	"github.com/gsk148/gophermart/internal/utils"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("UserID: failed to read from body %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		h.logger.Error("UserID: failed to unmarshal body %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userDB, err := h.repository.GetUserByLogin(user.Login)
	if !utils.CheckHashAndPassword(userDB.Password, user.Password) {
		h.logger.Error(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil && err != errors.ErrNoDBResult {
		h.logger.Error("failed to login, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = auth.GenerateCookie(w, userDB.ID)
	if err != nil {
		h.logger.Error("Failed to generate cookie, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Register: failed to read from body %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		h.logger.Error("Register: failed to unmarshal body %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Посмотрим, что юзера с таким логином нет
	userInDB, err := h.repository.GetUserByLogin(user.Login)
	switch {
	case err == nil && userInDB.Login != "":
		h.logger.Info("Register: user with provided login %s exists", user.Login)
		w.WriteHeader(http.StatusConflict)
		return
	case err == errors.ErrNoDBResult:
		cryptedPsw, err := utils.HashString(user.Password)
		if err != nil {
			h.logger.Error("Register: failed to encrypt password")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID, err := h.repository.Register(user.Login, cryptedPsw)
		if err != nil {
			h.logger.Error("Register: failed while registering in storage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = auth.GenerateCookie(w, userID)
		if err != nil {
			h.logger.Error("Failed to generate cookie, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
