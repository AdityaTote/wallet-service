package models

import (
	"github.com/google/uuid"
)

type WalletRequest struct {
	TxnId uuid.UUID `json:"txn_id" validate:"required"`
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type Wallet struct {
	ID uuid.UUID
}

type WalletServiceParams struct {
	WalletRequest
	UserId uuid.UUID
	WalletId uuid.UUID
}

type WalletResponse struct {
	Message string `json:"message"`
	Balance int64 `json:"balance"`
}