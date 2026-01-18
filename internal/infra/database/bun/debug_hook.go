package bun

import (
	"context"
	"ichi-go/pkg/logger"
	"time"

	"github.com/uptrace/bun"
)

type DebugHook struct{}

func (h *DebugHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *DebugHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	logger.Debugf("Query: %s\nDuration: %v ms\n", event.Query, time.Since(event.StartTime).Milliseconds())
}
