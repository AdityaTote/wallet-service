package models

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidBody  = errors.New("invalid request body")
	ErrInvalidInput = errors.New("invalid input")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Wallet service specific errors
var (
	ErrDuplicateTransaction = &AppError{
		Err:        errors.New("transaction with this ID already exists"),
		Message:    "transaction with this ID already exists",
		StatusCode: http.StatusConflict,
	}
	ErrWalletNotFound = &AppError{
		Err:        errors.New("wallet not found"),
		Message:    "wallet not found",
		StatusCode: http.StatusNotFound,
	}
	ErrInsufficientBalance = &AppError{
		Err:        errors.New("insufficient balance"),
		Message:    "insufficient balance",
		StatusCode: http.StatusBadRequest,
	}
	ErrInvalidBalance = &AppError{
		Err:        errors.New("invalid balance type"),
		Message:    "invalid balance type",
		StatusCode: http.StatusInternalServerError,
	}
	ErrTransactionFailed = &AppError{
		Err:        errors.New("transaction failed"),
		Message:    "transaction failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrBalanceRetrievalFailed = &AppError{
		Err:        errors.New("failed to retrieve balance"),
		Message:    "failed to retrieve balance",
		StatusCode: http.StatusInternalServerError,
	}
)

// NewAppError creates a new AppError
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}