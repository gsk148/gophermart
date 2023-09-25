package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/errors"
)

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromToken(w, r)

	balance, err := h.repository.GetBalance(userID)
	if err != nil {
		switch err {
		case errors.ErrNoDBResult:
			{
				h.logger.Info("GetBalance: no balance for provided userID", userID)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		default:
			h.logger.Error("GetBalance: error while select to db", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal(balance)
	if err != nil {
		h.logger.Error("GetBalance: failed to marshal resp %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
