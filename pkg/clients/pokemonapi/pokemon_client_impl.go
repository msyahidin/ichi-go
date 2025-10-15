package pokemonapi

import (
	"context"
	"fmt"
	"ichi-go/config"
	"ichi-go/internal/infra/http"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"resty.dev/v3"
	"time"
)

type PokemonClientImpl struct {
	httpClient *resty.Client
}

func NewPokemonClientImpl() *PokemonClientImpl {
	httpClient := http.New()
	cfg := config.Get().PkgClient().PokemonAPI
	httpClient.SetBaseURL(cfg.BaseURL)
	httpClient.SetTimeout(time.Duration(cfg.Timeout))
	httpClient.SetRetryCount(cfg.RetryCount)

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
