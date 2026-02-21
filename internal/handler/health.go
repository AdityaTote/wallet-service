package handler

import (
	"net/http"

	"github.com/AdityaTote/wallet-service/internal/lib/utils"
	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/service"
)

type HealthHandler interface {
	CheckHealth(w http.ResponseWriter, r *http.Request)
}

type health struct {
	svc service.HealthService
}

func (h *health) CheckHealth(w http.ResponseWriter, r *http.Request) {
	data := h.svc.CheckHealth()
	utils.JSONWriter(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "service is healthy",
		Data: data,
	})
}