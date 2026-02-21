package handler

import (
	"net/http"

	"github.com/AdityaTote/wallet-service/internal/lib/utils"
	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/service"
	"github.com/AdityaTote/wallet-service/internal/validations"
	"github.com/rs/zerolog"
)

type AuthHandler interface {
	Signup(w http.ResponseWriter, r *http.Request)
	Signin(w http.ResponseWriter, r *http.Request)
}

type auth struct {
	svc service.AuthService
	log zerolog.Logger
}

func (h *auth) Signup(w http.ResponseWriter, r *http.Request) {
	input, err := validations.ValidateAuthInput(r, h.log)
	if err != nil {
		switch err {
		case models.ErrInvalidBody:
			utils.JSONWriter(w, http.StatusBadRequest, models.JSONResponse{
				Success: false,
				Message: err.Error(),
			})
		default:
			utils.JSONWriter(w, http.StatusUnprocessableEntity, models.JSONResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		return
	}

	user, err := h.svc.Signup(models.UserParams{
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		utils.JSONWriter(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	utils.SetCookie(w, "ssid", user.AccessToken, true)

	utils.JSONWriter(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}


func (h *auth) Signin(w http.ResponseWriter, r *http.Request) {
	input, err := validations.ValidateAuthInput(r, h.log)
	if err != nil {
		switch err {
		case models.ErrInvalidBody:
			utils.JSONWriter(w, http.StatusBadRequest, models.JSONResponse{
				Success: false,
				Message: err.Error(),
			})
		default:
			utils.JSONWriter(w, http.StatusUnprocessableEntity, models.JSONResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		return
	}

	user, err := h.svc.Signin(models.UserParams{
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		utils.JSONWriter(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	utils.SetCookie(w, "ssid", user.AccessToken, true)

	utils.JSONWriter(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "User logged in successfully",
		Data:    user,
	})
}