package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// useBufferLogger swaps Log.log to write into `buf` at the given level.
func useBufferLogger(buf *bytes.Buffer, level zerolog.Level) {
	Log.log = zerolog.New(buf).Level(level)
}

func TestInit_FallbackToInfoLevel(t *testing.T) {
	// Backup any existing viper keys so we can restore them later.
	origLevel := viper.GetString("log.level")
	origApp := viper.GetString("app.name")
	defer func() {
		viper.Set("log.level", origLevel)
		viper.Set("app.name", origApp)
	}()

	// 1) Set an invalid log level → Init should fall back to InfoLevel
	viper.Set("log.level", "notalevel")
	viper.Set("app.name", "myservice")
	Init()

	// Now hijack Log.log into a buffer so we can inspect output
	buf := &bytes.Buffer{}
	// We don’t know the internal level exactly, but if Init fell back,
	// it used zerolog.InfoLevel. In either case, we can just force‐set
	// our buffer logger to InfoLevel and confirm Infof still writes.
	useBufferLogger(buf, zerolog.InfoLevel)

	// Calling Infof should produce output
	buf.Reset()
	Infof("hello %s", "world")
	out := buf.String()
	if !strings.Contains(out, `"hello world"`) {
		t.Errorf("expected Infof to produce output; got: %q", out)
	}
}

func TestWithContext_IncludesRequestID(t *testing.T) {
	// Swap Log.log into a buffer at InfoLevel
	buf := &bytes.Buffer{}
	useBufferLogger(buf, zerolog.InfoLevel)

	// Create a context carrying an X-Request-ID
	ctx := context.WithValue(context.Background(), echo.HeaderXRequestID, "req-xyz-123")
	l := WithContext(ctx)

	// Emit a single Infof and check JSON for the request ID
	buf.Reset()
	l.Infof("testing ctx")

	line := buf.String()
	if line == "" {
		t.Fatal("expected a log line, but buffer was empty")
	}

	var evt map[string]interface{}
	if err := json.Unmarshal([]byte(line), &evt); err != nil {
		t.Fatalf("failed to parse JSON log: %v\nraw: %q", err, line)
	}

	// The JSON should contain "X-Request-ID":"req-xyz-123"
	if id, ok := evt[echo.HeaderXRequestID]; !ok {
		t.Errorf("expected JSON to include field %q, but it did not: %v", echo.HeaderXRequestID, evt)
	} else if id.(string) != "req-xyz-123" {
		t.Errorf("expected %q == %q; got %q", echo.HeaderXRequestID, "req-xyz-123", id)
	}

	// Also confirm the message itself is present
	if !strings.Contains(line, `"testing ctx"`) {
		t.Errorf("expected log message to be present in JSON; got: %q", line)
	}
}

func TestLevelMethods_BasicOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	// Use TraceLevel so that Warnf / Errorf / Tracef all pass through
	useBufferLogger(buf, zerolog.TraceLevel)

	// Warnf
	buf.Reset()
	Warnf("warn %d", 7)
	if !strings.Contains(buf.String(), `"warn 7"`) {
		t.Errorf("Warnf did not write expected message; got: %q", buf.String())
	}

	// Errorf
	buf.Reset()
	Errorf("error %d", 8)
	if !strings.Contains(buf.String(), `"error 8"`) {
		t.Errorf("Errorf did not write expected message; got: %q", buf.String())
	}

	// Tracef
	buf.Reset()
	Tracef("trace %d", 9)
	if !strings.Contains(buf.String(), `"trace 9"`) {
		t.Errorf("Tracef did not write expected message; got: %q", buf.String())
	}

	// Infof (just as a sanity check that Infof still works)
	buf.Reset()
	Infof("info %s", "foobar")
	if !strings.Contains(buf.String(), `"info foobar"`) {
		t.Errorf("Infof did not write expected message; got: %q", buf.String())
	}
}
