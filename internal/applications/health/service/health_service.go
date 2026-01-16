package service

import (
	"context"
	"time"

	"ichi-go/pkg/health"
)

type HealthService struct {
	checker   *health.AggregateChecker
	version   string
	startTime time.Time
}

func NewHealthService(checker *health.AggregateChecker, version string) *HealthService {
	return &HealthService{
		checker:   checker,
		version:   version,
		startTime: time.Now(),
	}
}

// GetLiveness returns basic liveness status (fast, no dependency checks)
func (s *HealthService) GetLiveness(ctx context.Context) health.HealthResponse {
	return health.HealthResponse{
		Status:    health.StatusHealthy,
		Version:   s.version,
		Uptime:    time.Since(s.startTime),
		Timestamp: time.Now(),
	}
}

// GetReadiness checks all dependencies and returns readiness status
func (s *HealthService) GetReadiness(ctx context.Context) health.HealthResponse {
	// Create timeout context for health checks (3 seconds max)
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	components := s.checker.CheckAll(checkCtx)

	// Determine overall status
	overallStatus := health.StatusHealthy
	for _, component := range components {
		if component.Status == health.StatusUnhealthy {
			overallStatus = health.StatusUnhealthy
			break
		}
	}

	return health.HealthResponse{
		Status:     overallStatus,
		Version:    s.version,
		Uptime:     time.Since(s.startTime),
		Timestamp:  time.Now(),
		Components: components,
	}
}

// GetHealth is an alias to GetLiveness for backward compatibility
func (s *HealthService) GetHealth(ctx context.Context) health.HealthResponse {
	return s.GetLiveness(ctx)
}
