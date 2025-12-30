package http

import (
	"net/http"
	"resty.dev/v3"
	"time"
)

type ClientOptions struct {
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	RetryMaxWait  time.Duration
	LoggerEnabled bool
}

// New creates HTTP client with options
func New(opts ClientOptions) *resty.Client {
	client := resty.New()

	client.SetTimeout(opts.Timeout).
		SetRetryCount(opts.RetryCount).
		SetRetryWaitTime(opts.RetryWaitTime).
		SetRetryMaxWaitTime(opts.RetryMaxWait).
		AddRetryConditions(retryWhenStatusCodeNotOk)

	if opts.BaseURL != "" {
		client.SetBaseURL(opts.BaseURL)
	}

	if opts.LoggerEnabled {
		client.Logger()
	}

	return client
}

// NewFromConfig creates client from ClientConfig (backward compatibility)
func NewFromConfig(cfg *ClientConfig) *resty.Client {
	return New(ClientOptions{
		Timeout:       time.Duration(cfg.Timeout) * time.Millisecond,
		RetryCount:    cfg.RetryCount,
		RetryWaitTime: time.Duration(cfg.RetryWaitTime) * time.Millisecond,
		RetryMaxWait:  time.Duration(cfg.RetryMaxWait) * time.Millisecond,
		LoggerEnabled: cfg.LoggerEnabled,
	})
}

func retryWhenStatusCodeNotOk(response *resty.Response, err error) bool {
	return response.StatusCode() != http.StatusOK
}
