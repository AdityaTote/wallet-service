package router

import (
	"github.com/AdityaTote/wallet-service/internal/handler"
	"github.com/go-chi/chi/v5"
)

func authRouter(h handler.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/signup", h.Auth().Signup)
	r.Post("/signin", h.Auth().Signin)

	return r
}