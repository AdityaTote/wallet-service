package handler

import (
	"errors"
	"net/http"

	"github.com/AdityaTote/wallet-service/internal/lib/utils"
	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/service"
	"github.com/AdityaTote/wallet-service/internal/validations"
	"github.com/rs/zerolog"
)

type WalletHandler interface {
	TopUp(w http.ResponseWriter, r *http.Request)
	Spend(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
}

type wallet struct {
	svc service.WalletService
	log zerolog.Logger
}

func (h *wallet) TopUp(w http.ResponseWriter, r *http.Request) {
	urs, ok := r.Context().Value("user").(models.User)
	if !ok {
		utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}
	
	input, err := validations.ValidateWalletInput(r, h.log)
	if err != nil {
		utils.JSONWriter(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	data, err := h.svc.TopUp(&models.WalletServiceParams{
		UserId: urs.Id,
		WalletId: urs.WalletId,
		WalletRequest: models.WalletRequest{
			TxnId: input.TxnId,
			Amount: input.Amount,
		},
	})
	if err != nil {
		h.log.Error().Err(err).Msg("failed to top-up wallet")
		
		var appErr *models.AppError
		if errors.As(err, &appErr) {
			utils.JSONWriter(w, appErr.StatusCode, models.JSONResponse{
				Success: false,
				Message: appErr.Message,
			})
			return
		}
		
		utils.JSONWriter(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	message := "wallet topped up successfully"
	if data.Message != "" {
		message = data.Message
	}

	utils.JSONWriter(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: message,
		Data:    data.Balance,
	})
}

func (h *wallet) Spend(w http.ResponseWriter, r *http.Request) {
	urs, ok := r.Context().Value("user").(models.User)
	if !ok {
		utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}
	
	input, err := validations.ValidateWalletInput(r, h.log)
	if err != nil {
		utils.JSONWriter(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	data, err := h.svc.Spend(&models.WalletServiceParams{
		UserId: urs.Id,
		WalletId: urs.WalletId,
		WalletRequest: models.WalletRequest{
			TxnId: input.TxnId,
			Amount: input.Amount,
		},
	})
	if err != nil {
		h.log.Error().Err(err).Msg("failed to spend wallet")
		
		var appErr *models.AppError
		if errors.As(err, &appErr) {
			utils.JSONWriter(w, appErr.StatusCode, models.JSONResponse{
				Success: false,
				Message: appErr.Message,
			})
			return
		}
		
		utils.JSONWriter(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	message := "wallet spend up successfully"
	if data.Message != "" {
		message = data.Message
	}

	utils.JSONWriter(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: message,
		Data:    data.Balance,
	})
}

func (h *wallet) GetBalance(w http.ResponseWriter, r *http.Request) {
	urs, ok := r.Context().Value("user").(models.User)
	if !ok {
		utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}
	
	data, err := h.svc.Balance(urs.Id, urs.WalletId)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get wallet balance")

		var appErr *models.AppError
		if errors.As(err, &appErr) {
			utils.JSONWriter(w, appErr.StatusCode, models.JSONResponse{
				Success: false,
				Message: appErr.Message,
			})
			return
		}

		utils.JSONWriter(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	utils.JSONWriter(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "wallet balance retrieved successfully",
		Data:    data,
	})
}