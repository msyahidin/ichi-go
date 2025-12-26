package user

import (
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"ichi-go/config"
	user "ichi-go/internal/applications/user/controller"
	userRepo "ichi-go/internal/applications/user/repository"
	userService "ichi-go/internal/applications/user/service"
	"ichi-go/pkg/clients/pokemonapi"

	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// RegisterProviders registers all user domain dependencies
func RegisterProviders(injector do.Injector) {
	do.Provide(injector, ProvidePokemonClient)
	do.Provide(injector, ProvideUserRepository)
	do.Provide(injector, ProvideUserService)
	do.Provide(injector, ProvideUserController)
}

func ProvidePokemonClient(i do.Injector) (pokemonapi.PokemonClient, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return pokemonapi.NewPokemonClientImpl(cfg), nil
}

func ProvideUserRepository(i do.Injector) (*userRepo.RepositoryImpl, error) {
	db := do.MustInvoke[*bun.DB](i)
	return userRepo.NewUserRepository(db), nil
}

func ProvideUserService(i do.Injector) (*userService.ServiceImpl, error) {
	repo := do.MustInvoke[*userRepo.RepositoryImpl](i)
	cacheClient := do.MustInvoke[*redis.Client](i)
	cacheImpl := cache.NewCache(cacheClient)
	pokeClient := do.MustInvoke[pokemonapi.PokemonClient](i)
	// Queue producer is optional
	var producer rabbitmq.MessageProducer
	if conn, err := do.Invoke[*rabbitmq.Connection](i); err == nil && conn != nil {
		cfg := do.MustInvoke[*config.Config](i)
		// Create producer from connection
		if p, err := rabbitmq.NewProducer(conn, cfg.Queue().RabbitMQ); err == nil {
			producer = p
		}
	}

	return userService.NewUserService(repo, cacheImpl, pokeClient, producer), nil
}

func ProvideUserController(i do.Injector) (*user.UserController, error) {
	svc := do.MustInvoke[*userService.ServiceImpl](i)
	return user.NewUserController(svc), nil
}
