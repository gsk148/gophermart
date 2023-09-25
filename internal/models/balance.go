package models

type GetBalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WithdrawBalanceRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
