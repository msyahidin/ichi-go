package auth

import (
	"ichi-go/config"
	authController "ichi-go/internal/applications/auth/controller"
	authService "ichi-go/internal/applications/auth/service"
	userRepo "ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/authenticator"

	"github.com/samber/do/v2"
)

// RegisterProviders registers all auth domain dependencies
func RegisterProviders(injector do.Injector) {
	do.Provide(injector, ProvideAuthService)
	do.Provide(injector, ProvideAuthController)
}

// ProvideAuthService provides auth service instance
func ProvideAuthService(i do.Injector) (*authService.ServiceImpl, error) {
	userRepository := do.MustInvoke[*userRepo.RepositoryImpl](i)
	cfg := do.MustInvoke[*config.Config](i)

	// Create JWT authenticator
	jwtAuth := authenticator.NewJWTAuthenticator(cfg.Auth().JWT)
	// Queue producer is optional
	var producer rabbitmq.MessageProducer
	if conn, err := do.Invoke[*rabbitmq.Connection](i); err == nil && conn != nil {
		// Create producer from connection
		if p, err := rabbitmq.NewProducer(conn, cfg.Queue().RabbitMQ); err == nil {
			producer = p
		}
	}

	return authService.NewAuthService(userRepository, jwtAuth, producer), nil
}

// ProvideAuthController provides auth controller instance
func ProvideAuthController(i do.Injector) (*authController.AuthController, error) {
	svc := do.MustInvoke[*authService.ServiceImpl](i)
	return authController.NewAuthController(svc), nil
}
