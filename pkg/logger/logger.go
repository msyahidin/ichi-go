package logger

import (
	"github.com/rs/zerolog"
	"ichi-go/config"
	"os"
	"time"
)

var Log zerolog.Logger

func Init() {
	logLevel := config.Cfg.Log.Level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339

	Log = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(level)
}
