package clients

import pokeConfig "ichi-go/pkg/clients/pokemonapi/config"

type PkgClient struct {
	PokemonAPI pokeConfig.Config `mapstructure:"pokemon_api"`
}
