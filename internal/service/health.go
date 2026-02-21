package service

import (
	"context"
	"time"

	"github.com/AdityaTote/wallet-service/internal/server"
	"github.com/rs/zerolog"
)

type HealthServiceResponse struct {
	Status      string                            `json:"status"`
	Timestamp   time.Time                         `json:"timestamp"`
	Checks      map[string]map[string]interface{} `json:"checks"`
}

type HealthService interface {
	CheckHealth() HealthServiceResponse
}

type healthService struct {
	ctx   context.Context
	log  zerolog.Logger
	srv *server.Server
}

func (s *healthService) CheckHealth() HealthServiceResponse {
	response := HealthServiceResponse{
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Checks:      map[string]map[string]interface{}{},
	}

	check := response.Checks
	isHealth := true

	dbStart := time.Now()
	if err := s.srv.Database.Pool.Ping(s.ctx); err != nil {
		check["database"] = map[string]interface{}{
			"status":        "unhealthy",
			"response_time": time.Since(dbStart).String(),
		}
		isHealth = false
		s.log.Info().Err(err).Dur("response_time", time.Since(dbStart)).Msg("database health check failed")
	} else {
		check["database"] = map[string]interface{}{
			"status":        "healthy",
			"response_time": time.Since(dbStart).String(),
		}
		s.log.Info().Dur("response_time", time.Since(dbStart)).Msg("database health check passed")
	}

	if !isHealth {
		response.Status = "unhealthy"
	}

	return response
}