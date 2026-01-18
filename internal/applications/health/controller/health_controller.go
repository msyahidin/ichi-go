package controller

import (
	"net/http"

	"ichi-go/internal/applications/health/service"
	"ichi-go/pkg/health"

	"github.com/labstack/echo/v4"
)

var Domain = "health"

type HealthController struct {
	service *service.HealthService
}

func NewHealthController(service *service.HealthService) *HealthController {
	return &HealthController{service: service}
}

// RegisterRoutes registers health check routes (NO versioning, NO auth)
func (c *HealthController) RegisterRoutes(e *echo.Echo, serviceName string) {
	// Health check endpoints - bypass auth and versioning
	group := e.Group("/" + serviceName + "/api/" + Domain)
	group.GET("", c.GetHealth)
	group.GET("/live", c.GetLiveness)
	group.GET("/ready", c.GetReadiness)
}

// GetHealth godoc
//
//	@Summary		Basic health check
//	@Description	Returns basic service health status and uptime
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	health.HealthResponse
//	@Router			/health [get]
func (c *HealthController) GetHealth(ctx echo.Context) error {
	response := c.service.GetHealth(ctx.Request().Context())
	return ctx.JSON(http.StatusOK, response)
}

// GetLiveness godoc
//
//	@Summary		Kubernetes liveness probe
//	@Description	Fast health check for Kubernetes liveness probe
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	health.HealthResponse
//	@Router			/health/live [get]
func (c *HealthController) GetLiveness(ctx echo.Context) error {
	response := c.service.GetLiveness(ctx.Request().Context())
	return ctx.JSON(http.StatusOK, response)
}

// GetReadiness godoc
//
//	@Summary		Kubernetes readiness probe
//	@Description	Checks all dependencies (database, redis, rabbitmq) for readiness
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	health.HealthResponse	"All dependencies healthy"
//	@Failure		503	{object}	health.HealthResponse	"One or more dependencies unhealthy"
//	@Router			/health/ready [get]
func (c *HealthController) GetReadiness(ctx echo.Context) error {
	response := c.service.GetReadiness(ctx.Request().Context())

	statusCode := http.StatusOK
	if response.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	return ctx.JSON(statusCode, response)
}
