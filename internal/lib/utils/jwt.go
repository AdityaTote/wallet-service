package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	UserID string `json:"uid"`
	WalletID uuid.UUID `json:"wid"`
	jwt.RegisteredClaims
}

var (
	accessTTL  = 24 * time.Hour
)

func GenerateAccessToken(accessSecret []byte, userID uuid.UUID, walletId uuid.UUID) (string, error) {
	claims := AccessClaims{
		UserID: userID.String(),
		WalletID: walletId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wallet-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}

func ParseAccessToken(accessSecret []byte, tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&AccessClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return accessSecret, nil
		},
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token.Claims.(*AccessClaims), nil
}