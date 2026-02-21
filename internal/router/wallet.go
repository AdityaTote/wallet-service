package router

import (
	"github.com/AdityaTote/wallet-service/internal/handler"
	"github.com/AdityaTote/wallet-service/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func walletRouter(h handler.Handlers, authMiddleware *middleware.AuthMiddleware) *chi.Mux {
	r := chi.NewRouter()

	// apply auth middleware to all wallet routes
	r.Use(authMiddleware.Middleware)

	r.Get("/balance", h.Wallet().GetBalance)
	r.Post("/topup", h.Wallet().TopUp)
	r.Post("/spend", h.Wallet().Spend)

	return r
}