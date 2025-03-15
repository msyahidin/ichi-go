package pokemon_api

import (
	"context"
	"ichi-go/internal/infra/external/pokemon_api/dto"
)

type PokemonClient interface {
	GetDetail(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error)
}
