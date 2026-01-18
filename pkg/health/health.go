package health

import (
	"context"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentHealth represents the health of a single dependency
type ComponentHealth struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Latency   time.Duration `json:"latency_ms,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// HealthResponse represents the overall health status
type HealthResponse struct {
	Status     Status                     `json:"status"`
	Version    string                     `json:"version"`
	Uptime     time.Duration              `json:"uptime_seconds"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
}

// Checker is the interface for checking component health
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}
