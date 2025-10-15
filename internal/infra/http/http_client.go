package http

import (
	"ichi-go/config"
	"net/http"
	"resty.dev/v3"
	"time"
)

func New() *resty.Client {
	client := resty.New()
	defer client.Close()
	cfg := config.Get()

	client.SetTimeout(time.Duration(cfg.HttpClient().Timeout) * time.Millisecond).
		SetRetryCount(cfg.HttpClient().RetryCount).
		SetRetryWaitTime(time.Duration(cfg.HttpClient().RetryWaitTime) * time.Millisecond).
		SetRetryMaxWaitTime(time.Duration(cfg.HttpClient().RetryMaxWait) * time.Millisecond).
		AddRetryConditions(retryWhenStatusCodeNotOk)

	if cfg.HttpClient().LoggerEnabled {
		client.Logger()
	}
	return client
}

func retryWhenStatusCodeNotOk(response *resty.Response, err error) bool {
	return response.StatusCode() != http.StatusOK
}
