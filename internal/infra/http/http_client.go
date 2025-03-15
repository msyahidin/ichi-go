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

	client.SetTimeout(time.Duration(config.HttpClient().Timeout) * time.Millisecond).
		SetRetryCount(config.HttpClient().RetryCount).
		SetRetryWaitTime(time.Duration(config.HttpClient().RetryWaitTime) * time.Millisecond).
		SetRetryMaxWaitTime(time.Duration(config.HttpClient().RetryMaxWait) * time.Millisecond).
		AddRetryConditions(retryWhenStatusCodeNotOk)

	if config.HttpClient().LoggerEnabled {
		client.Logger()
	}
	return client
}

func retryWhenStatusCodeNotOk(response *resty.Response, err error) bool {
	return response.StatusCode() != http.StatusOK
}
