package http

import (
	configHttp "ichi-go/config/http"
	"net/http"
	"time"

	"resty.dev/v3"
)

func New(cfg *configHttp.ClientConfig) *resty.Client {
	client := resty.New()
	defer client.Close()

	client.SetTimeout(time.Duration(cfg.Timeout) * time.Millisecond).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(time.Duration(cfg.RetryWaitTime) * time.Millisecond).
		SetRetryMaxWaitTime(time.Duration(cfg.RetryMaxWait) * time.Millisecond).
		AddRetryConditions(retryWhenStatusCodeNotOk)

	if cfg.LoggerEnabled {
		client.Logger()
	}
	return client
}

func retryWhenStatusCodeNotOk(response *resty.Response, err error) bool {
	return response.StatusCode() != http.StatusOK
}
