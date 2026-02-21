package handler

import (
	"github.com/AdityaTote/wallet-service/internal/service"
	"github.com/rs/zerolog"
)


type Handlers interface {
	Health() HealthHandler
	Auth() AuthHandler
	Wallet() WalletHandler
}

type handlers struct {
	svc service.Services
	log zerolog.Logger
}

func New(svc service.Services, log zerolog.Logger) Handlers {
	return &handlers{
		svc: svc,
		log: log,
	}
}

func (h *handlers) Health() HealthHandler {
	return &health{
		svc: h.svc.Health(),
	}
}

func (h *handlers) Auth() AuthHandler {
	return &auth{
		svc: h.svc.Auth(),
		log: h.log,
	}
}

func (h *handlers) Wallet() WalletHandler {
	return &wallet{
		svc: h.svc.Wallet(),
		log: h.log,
	}
}