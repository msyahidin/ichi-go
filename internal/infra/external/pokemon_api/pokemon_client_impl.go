package pokemon_api

import (
	"context"
	"fmt"
	"ichi-go/config"
	"ichi-go/internal/infra/external/pokemon_api/dto"
	"ichi-go/internal/infra/http"
	"resty.dev/v3"
	"time"
)

type PokemonClientImpl struct {
	httpClient *resty.Client
}

func NewPokemonClientImpl() *PokemonClientImpl {
	httpClient := http.New()
	httpClient.SetBaseURL(config.External().PokemonAPI.BaseURL)
	httpClient.SetTimeout(time.Duration(config.External().PokemonAPI.Timeout))
	httpClient.SetRetryCount(config.External().PokemonAPI.RetryCount)

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
