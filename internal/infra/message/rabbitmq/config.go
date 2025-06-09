package rabbitmq

import (
	"crypto/tls"
	"github.com/cenkalti/backoff/v4"
)

type Config struct {
	URI              string
	ReconnectBackoff backoff.BackOff
	TLSConfig        *tls.Config
}
