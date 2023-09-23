package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/errors"
	"github.com/gsk148/gophermart/internal/models"
)

func (h *Handler) DeductPoints(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromToken(w, r)

	var withdrawRequest models.WithdrawBalanceRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("DeductPoints: failed while read body", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	buf := bytes.NewBuffer(body)
	err = json.NewDecoder(buf).Decode(&withdrawRequest)
	if err != nil {
		h.logger.Error("DeductPoints: failed while decode", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err = goluhn.Validate(withdrawRequest.Order); err != nil {
		h.logger.Info("Provided order num invalid")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	balance, err := h.repository.GetBalance(userID)
	if err != nil {
		h.logger.Error("GetBalance: failed, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if balance.Current < withdrawRequest.Sum {
		h.logger.Info("Withdraw: balance not enough")
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	// если нет такого заказа, мы его создаем
	_, err = h.repository.GetOrderByNumber(withdrawRequest.Order)
	if err == errors.ErrNoDBResult {
		h.logger.Info("DeductPoints: provided order with num %s not exist, creating", withdrawRequest.Order)
		h.repository.LoadOrder(withdrawRequest.Order, userID)
	}

	// cписываем баллы
	err = h.repository.DeductPoints(withdrawRequest, userID, withdrawRequest.Order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Info("Order received to withdraw loyalty", withdrawRequest.Order)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := auth.GetUserIDFromToken(w, r)

	withdrawals, err := h.repository.GetWithdrawals(userID)
	if err != nil {
		switch err {
		case errors.ErrNoDBResult:
			{
				h.logger.Info("No withdrawals for provided user", userID)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		default:
			{
				h.logger.Error("Failed to get withdrawals", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	h.logger.Info("%v user withdrawals are %v", userID, withdrawals)
	resp, err := json.Marshal(withdrawals)
	if err != nil {
		h.logger.Error("Failed to marshal get withdrawals", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resp)
	w.WriteHeader(http.StatusOK)
}
