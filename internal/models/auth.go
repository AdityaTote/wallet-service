package models

import "github.com/google/uuid"

type User struct {
	Id uuid.UUID
	WalletId uuid.UUID
}

type UserParams struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	Id uuid.UUID `json:"id"`
	Username string `json:"username"`
	WalletId    *uuid.UUID `json:"wallet_id,omitempty"`
	Balance     *int64     `json:"balance,omitempty"`
	AccessToken string `json:"access_token"`
}