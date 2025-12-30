package pokemonapi

import (
	"context"
	"fmt"
	"ichi-go/pkg/clients/pokemonapi/dto"
	pkgHttp "ichi-go/pkg/http"
	"resty.dev/v3"
)

type PokemonClientImpl struct {
	httpClient *resty.Client
}

// NewPokemonClient creates Pokemon client with merged configuration
func NewPokemonClient(pokemonCfg Config, httpDefaults HTTPClientDefaults) *PokemonClientImpl {
	opts := pokemonCfg.MergeWithDefaults(httpDefaults)

	httpClient := pkgHttp.New(pkgHttp.ClientOptions{
		BaseURL:       opts.BaseURL,
		Timeout:       opts.Timeout,
		RetryCount:    opts.RetryCount,
		RetryWaitTime: opts.RetryWaitTime,
		RetryMaxWait:  opts.RetryMaxWait,
		LoggerEnabled: opts.LoggerEnabled,
	})

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
