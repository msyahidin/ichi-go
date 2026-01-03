package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/samber/oops/loggers/zerolog"
	"github.com/spf13/viper"
)

type Logger struct {
	zerolog.Logger
}

var (
	once     sync.Once
	instance *Logger
)

type CallerWrapperHook struct{}

func (h CallerWrapperHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	_, file, line, ok := runtime.Caller(5)
	if ok {
		e.Str("caller", fmt.Sprintf("%s/%s:%d", filepath.Dir(file), filepath.Base(file), line))
	}
}

func GetInstance() *Logger {
	once.Do(func() {
		pretty := viper.GetBool("log.pretty")
		logLevel := viper.GetString("log.level")
		level, err := zerolog.ParseLevel(logLevel)

		if err != nil {
			level = zerolog.InfoLevel
		}
		zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
		zerolog.ErrorMarshalFunc = oopszerolog.OopsMarshalFunc
		debug := viper.GetBool("app.debug")
		if debug {
			level = zerolog.DebugLevel
		}

		zerolog.TimeFieldFormat = "2006-01-02 15:04:05.00007Z07:00"
		wd, _ := os.Getwd()
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			file = strings.TrimPrefix(file, wd+"/")
			return fmt.Sprintf("%s:%d", file, line)
		}
		var output io.Writer
		output = os.Stdout
		if level == zerolog.DebugLevel {
			// zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
			zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
		}
		if pretty {
			output = zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false,
				TimeFormat: time.DateTime, // "2006-01-02 15:04:05"
				FormatTimestamp: func(i interface{}) string {
					return fmt.Sprintf("\x1b[36m%s\x1b[0m", i)
				},
				FormatLevel: func(i interface{}) string {
					fLevel := strings.ToUpper(fmt.Sprintf("%s", i))
					switch fLevel {
					case "ERROR":
						return fmt.Sprintf("\x1b[31m| %-5s |\x1b[0m", fLevel)
					case "WARN":
						return fmt.Sprintf("\x1b[33m| %-5s |\x1b[0m", fLevel)
					case "DEBUG":
						return fmt.Sprintf("\x1b[94m| %-5s |\x1b[0m", fLevel)
					case "INFO":
						return fmt.Sprintf("\x1b[32m| %-5s |\x1b[0m", fLevel)
					default:
						return fmt.Sprintf("| %-5s|", fLevel)
					}
				},
				FormatMessage: func(i interface{}) string {
					return fmt.Sprintf("%s", i)
				},
				FormatFieldName: func(i interface{}) string {
					return fmt.Sprintf("%s=", i)
				},
				FormatFieldValue: func(i interface{}) string {
					return fmt.Sprintf("%v", i)
				}}
		}

		instance = &Logger{zerolog.New(output).
			With().
			Timestamp().
			Str("service", viper.GetString("app.name")).
			Logger().
			Level(level)}
	})
	return instance
}

func RequestLogging(eCtx echo.Context, format string, v ...interface{}) {
	WithRequestStamp(eCtx).Info().
		Msgf(format, v...)
}

func inner() error {
	return errors.New("something went wrong")
}

func middle() error {
	err := inner()
	if err != nil {
		return err
	}
	return nil
}

func outer() error {
	err := middle()
	if err != nil {
		return err
	}
	return nil
}

func WithContext(ctx context.Context) *Logger {
	baseInstance := GetInstance()
	contextualLogger := baseInstance.With().Logger()

	if reqID, ok := ctx.Value(echo.HeaderXRequestID).(string); ok {
		contextualLogger = contextualLogger.With().Str(echo.HeaderXRequestID, reqID).Logger()
	}
	return &Logger{contextualLogger}
}

func WithRequestStamp(eCtx echo.Context) *Logger {
	baseInstance := GetInstance()
	contextualLogger := baseInstance.With().
		Str("method", eCtx.Request().Method).
		Str("path", eCtx.Request().URL.Path).
		Int("status", eCtx.Response().Status).
		Str("ip", eCtx.RealIP()).
		Str("user_agent", eCtx.Request().UserAgent()).
		Str(echo.HeaderXRequestID, eCtx.Response().Header().Get(echo.HeaderXRequestID)).
		Str("domain", eCtx.Request().Header.Get("domain")).Logger()
	return &Logger{contextualLogger}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Info().Msgf(format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	logger := l.Logger.Hook(CallerWrapperHook{})
	logger.
		Debug().
		Msgf(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logger.Warn().Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	logger := l.Logger.Hook(CallerWrapperHook{})
	logger.Error().Msgf(format, v...)
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	l.Logger.Trace().Msgf(format, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	logger := l.Logger.Hook(CallerWrapperHook{})
	logger.Panic().Msgf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	logger := l.Logger.Hook(CallerWrapperHook{})
	logger.Fatal().Msgf(format, v...)
}

func (l *Logger) Trace(v ...interface{}) {
	logger := l.Logger.Hook(CallerWrapperHook{})
	logger.Trace().Msgf("%v", v...)
}

func (l *Logger) Panic(v ...interface{}) {
	l.Logger.Panic().Msgf("%v", v...)
}

func (l *Logger) Error(err error, v ...interface{}) {
	l.Logger.Error().
		Stack().
		Err(err).
		Msgf("%v", v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Logger.Warn().Msgf("%v", v...)
}

func Infof(format string, v ...interface{}) {
	GetInstance().Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	GetInstance().Debugf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	GetInstance().Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	GetInstance().Errorf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	GetInstance().Panicf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	GetInstance().Fatalf(format, v...)
}

func Tracef(format string, v ...interface{}) {
	GetInstance().Tracef(format, v...)
}

func DebugContext(ctx context.Context, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Debug().Msgf("%v", v...)
}

func DebugContextf(ctx context.Context, format string, v ...interface{}) {
	withContext := WithContext(ctx)
	withContext.Debugf(format, v...)
}
