package pokemonapi

import (
	"context"
	"fmt"
	"ichi-go/config"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"ichi-go/pkg/http"
	"resty.dev/v3"
	"time"
)

type PokemonClientImpl struct {
	httpClient *resty.Client
}

func NewPokemonClientImpl(cfg *config.Config) *PokemonClientImpl {
	httpClient := http.New(cfg.HttpClient())

	httpClient.SetBaseURL(cfg.PkgClient().PokemonAPI.BaseURL)
	httpClient.SetTimeout(time.Duration(cfg.PkgClient().PokemonAPI.Timeout))
	httpClient.SetRetryCount(cfg.PkgClient().PokemonAPI.RetryCount)

	return &PokemonClientImpl{
		httpClient: httpClient,
	}
}

func (p *PokemonClientImpl) GetDetail(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error) {
	var getDto dto.PokemonGetResponseDto
	resp, err := p.httpClient.R().
		SetContext(ctx).
		SetResult(&getDto).
		Get("pokemon/" + name)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error response: %v", resp)
	}
	return &getDto, nil
}
