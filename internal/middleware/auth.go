package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AdityaTote/wallet-service/internal/lib/utils"
	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/repository"
	"github.com/AdityaTote/wallet-service/internal/server"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AuthMiddleware struct {
	server *server.Server
	repo   *repository.Repository
	log    zerolog.Logger
}

func NewAuthMiddleware(server *server.Server, repo *repository.Repository, log zerolog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		server: server,
		repo:   repo,
		log:    log,
	}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := utils.GetCookieValue(r, "ssid")

		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
					Success: false,
					Message: "unauthorized",
				})
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
					Success: false,
					Message: "unauthorized",
				})
				return
			}
			token = parts[1]
		}

		// check token
		if token == "" {
			utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
				Success: false,
				Message: "unauthorized",
			})
			return
		}

		// verify access token
		user, err := utils.ParseAccessToken([]byte(m.server.Config.JWTSecret), token)
		if err != nil {
			utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
				Success: false,
				Message: "unauthorized",
			})
			return
		}

		// get queries
		query := m.repo.Queries()

		// parse user ID
		userID, err := uuid.Parse(user.UserID)
		if err != nil {
			m.log.Error().Err(err).Msg("invalid user ID in token")
			utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
				Success: false,
				Message: "unauthorized",
			})
			return
		}

		// check if user exists
		u, err := query.GetUserById(r.Context(), userID)
		if err != nil {
			m.log.Error().Err(err).Msg("user not found")
			utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
				Success: false,
				Message: "unauthorized",
			})
			return
		}

		// check if user has a wallet
		wallet, err := query.GetWalletByOwner(r.Context(), u.ID)
		if err != nil {
			m.log.Error().Err(err).Msg("wallet not found for user")
			utils.JSONWriter(w, http.StatusUnauthorized, models.JSONResponse{
				Success: false,
				Message: "unauthorized",
			})
			return
		}

		// add user info to context
		ctx := context.WithValue(r.Context(), "user", models.User{
			Id: u.ID,
			WalletId: wallet.ID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}