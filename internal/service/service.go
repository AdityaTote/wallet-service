package service

import (
	"context"

	"github.com/AdityaTote/wallet-service/internal/config"
	"github.com/AdityaTote/wallet-service/internal/repository"
	"github.com/AdityaTote/wallet-service/internal/server"
	"github.com/rs/zerolog"
)

type Services interface {
	Health() HealthService
	Auth() AuthService
	Wallet() WalletService
}

type service struct {
	ctx context.Context
	config *config.Config
	log zerolog.Logger
	srv *server.Server
	repo *repository.Repository
}

func New(ctx context.Context, cfg *config.Config, log zerolog.Logger, srv *server.Server, repo *repository.Repository) Services {
	return &service{
		ctx: ctx,
		config: cfg,
		log: log,
		srv: srv,
		repo: repo,
	}
}

func (s *service) Health() HealthService  {
	return &healthService{
		ctx: s.ctx,
		log: s.log,
		srv: s.srv,
	}
}

func (s *service) Auth() AuthService  {
	return &authService{
		ctx: s.ctx,
		cfg: s.config,
		log: s.log,
		repo: s.repo.Queries(),
	}
}

func (s *service) Wallet() WalletService  {
	return &walletService{
		ctx: s.ctx,
		log: s.log,
		repo: *s.repo,
	}
}