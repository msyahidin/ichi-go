package logger

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"time"
)

var Log zerolog.Logger

func Init() {
	logLevel := viper.GetString("log.level")
	debug := viper.Get("app.debug")
	if debug == true {
		logLevel = "debug"
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339

	Log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", viper.GetString("app.name")).
		Logger().
		Level(level)

	if level == zerolog.DebugLevel {
		Log = Log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
}

func WithContext(ctx context.Context) zerolog.Logger {
	if reqID, ok := ctx.Value(echo.HeaderXRequestID).(string); ok {
		return Log.With().Str(echo.HeaderXRequestID, reqID).Logger()
	}
	return Log
}

func Printf(format string, v ...interface{}) {
	Log.Info().Msgf(format, v...)
}

func Trace(v ...interface{}) {
	Log.Trace().Msgf("%v", v...)
}

func Panic(v ...interface{}) {
	Log.Panic().Msgf("%v", v...)
}

func Infof(format string, v ...interface{}) {
	Log.Info().Msgf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	Log.Debug().Msgf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	Log.Warn().Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	Log.Error().Msgf(format, v...)
}

func Tracef(format string, v ...interface{}) {
	Log.Trace().Msgf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	Log.Panic().Msgf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	Log.Fatal().Msgf(format, v...)
}

func PrintContextln(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Info().Msgf("%v", v...)
}

func DebugContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Debug().Msgf("%v", v...)
}

func DebugContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Debug().Msgf(format, v...)
}

func InfoContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Info().Msgf("%v", v...)
}

func InfoContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Info().Msgf(format, v...)
}

func WarnContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Warn().Msgf("%v", v...)
}

func WarnContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Warn().Msgf(format, v...)
}

func ErrorContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Error().Msgf("%v", v...)
}

func ErrorContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Error().Msgf(format, v...)
}

func TraceContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Trace().Msgf("%v", v...)
}

func TraceContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Trace().Msgf(format, v...)
}

func PanicContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Panic().Msgf("%v", v...)
}

func PanicContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Panic().Msgf(format, v...)
}

func FatalContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Fatal().Msgf("%v", v...)
}

func FatalContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Fatal().Msgf(format, v...)
}
