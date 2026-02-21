package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdityaTote/wallet-service/internal/config"
	"github.com/AdityaTote/wallet-service/internal/database"
)

type Server struct {
	Config *config.Config
	Database *database.Database
	httpServer *http.Server
}

func New(cfg *config.Config, db *database.Database) (*Server) {
	return &Server{
		Config: cfg,
		Database: db,
	}
}

func (s *Server) SetUpHttpServer(h http.Handler) {
	s.httpServer = &http.Server{
		Addr: ":" + s.Config.Port,
		Handler: h,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return fmt.Errorf("Http server not initiated")
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.Database.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	if err := s.httpServer.Close(); err != nil {
		return fmt.Errorf("failed to close http server: %w", err)
	}

	return nil
}