package router

import (
	"github.com/AdityaTote/wallet-service/internal/handler"
	"github.com/AdityaTote/wallet-service/internal/middleware"
	"github.com/AdityaTote/wallet-service/internal/repository"
	"github.com/AdityaTote/wallet-service/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func New(h handler.Handlers, srv *server.Server, repo *repository.Repository, log zerolog.Logger) *chi.Mux {
	// initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(srv, repo, log)
	router := chi.NewRouter()
	router.Mount("/api", apiRoutes(h, authMiddleware))
	return router
}

func apiRoutes(h handler.Handlers, authMiddleware *middleware.AuthMiddleware) *chi.Mux {
	r := chi.NewRouter()
	
	// routes
	r.Get("/health", h.Health().CheckHealth)
	r.Mount("/auth", authRouter(h))
	r.Mount("/wallet", walletRouter(h, authMiddleware))


	return r
}