package logger

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Logger struct {
	log zerolog.Logger
}

var Log Logger
var Debug = false

func Init(debug bool) {
	Debug = debug
	logLevel := viper.GetString("log.level")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339

	Log.log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", viper.GetString("app.name")).
		Logger().
		Level(level)

}

func (l Logger) GetLogger() zerolog.Logger {
	return l.log
}

func WithContext(ctx context.Context) Logger {
	if reqID, ok := ctx.Value(echo.HeaderXRequestID).(string); ok {
		return Logger{
			log: Log.log.With().Str(echo.HeaderXRequestID, reqID).Logger(),
		}
	}
	return Log
}

func (l Logger) Infof(format string, v ...interface{}) {
	l.log.Info().Msgf(format, v...)
}

func (l Logger) Debugf(format string, v ...interface{}) {
	if Debug {
		l.log.Debug().Msgf(format, v...)
	}
}

func (l Logger) Warnf(format string, v ...interface{}) {
	l.log.Warn().Msgf(format, v...)
}

func (l Logger) Errorf(format string, v ...interface{}) {
	l.log.Error().Msgf(format, v...)
}

func (l Logger) Tracef(format string, v ...interface{}) {
	l.log.Trace().Msgf(format, v...)
}

func (l Logger) Panicf(format string, v ...interface{}) {
	l.log.Panic().Msgf(format, v...)
}

func (l Logger) Fatalf(format string, v ...interface{}) {
	l.log.Fatal().Msgf(format, v...)
}

func (l Logger) Trace(v ...interface{}) {
	l.log.Trace().Msgf("%v", v...)
}

func (l Logger) Panic(v ...interface{}) {
	l.log.Panic().Msgf("%v", v...)
}

func (l Logger) Info(v ...interface{}) {
	l.log.Info().Msgf("%v", v...)
}

func (l Logger) Debug(v ...interface{}) {
	if Debug {
		l.log.Debug().Msgf("%v", v...)
	}
}

func (l Logger) Warn(v ...interface{}) {
	l.log.Warn().Msgf("%v", v...)
}

func (l Logger) Error(v ...interface{}) {
	l.log.Error().Msgf("%v", v...)
}

func (l Logger) Fatal(v ...interface{}) {
	l.log.Fatal().Msgf("%v", v...)
}

func Printf(format string, v ...interface{}) {
	Log.log.Info().Msgf(format, v...)
}

func Trace(v ...interface{}) {
	Log.log.Trace().Msgf("%v", v...)
}

func Panic(v ...interface{}) {
	Log.log.Panic().Msgf("%v", v...)
}

func Infof(format string, v ...interface{}) {
	Log.log.Info().Msgf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	if Debug {
		Log.log.Debug().Msgf(format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	Log.log.Warn().Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	Log.log.Error().Msgf(format, v...)
}

func Tracef(format string, v ...interface{}) {
	Log.log.Trace().Msgf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	Log.log.Panic().Msgf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	Log.log.Fatal().Msgf(format, v...)
}

func PrintContextln(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Info().Msgf("%v", v...)
}

func DebugContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Debug().Msgf("%v", v...)
}

func DebugContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Debug().Msgf(format, v...)
}

func InfoContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Info().Msgf("%v", v...)
}

func InfoContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Info().Msgf(format, v...)
}

func WarnContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Warn().Msgf("%v", v...)
}

func WarnContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Warn().Msgf(format, v...)
}

func ErrorContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Error().Msgf("%v", v...)
}

func ErrorContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Error().Msgf(format, v...)
}

func TraceContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Trace().Msgf("%v", v...)
}

func TraceContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Trace().Msgf(format, v...)
}

func PanicContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Panic().Msgf("%v", v...)
}

func PanicContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Panic().Msgf(format, v...)
}

func FatalContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Fatal().Msgf("%v", v...)
}

func FatalContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.log.Fatal().Msgf(format, v...)
}
