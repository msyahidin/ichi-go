package pokemonapi

import (
	"context"
	"ichi-go/pkg/clients/pokemonapi/dto"
)

type PokemonClient interface {
	GetDetail(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error)
}
