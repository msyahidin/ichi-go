package notification

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"

	notifController "ichi-go/internal/applications/notification/controller"
	"ichi-go/pkg/authenticator"
)

// Register wires up the notification domain: DI providers, routes.
// Call this from cmd/server/rest_server.go alongside other domain registrations.
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	RegisterProviders(injector)

	ctrl := do.MustInvoke[*notifController.NotificationController](injector)
	ctrl.RegisterRoutes(e, serviceName, auth)
}
