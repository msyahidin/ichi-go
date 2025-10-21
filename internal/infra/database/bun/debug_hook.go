package bun

import (
	"context"
	"github.com/uptrace/bun"
	"ichi-go/pkg/logger"
	"time"
)

type DebugHook struct{}

func (h *DebugHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *DebugHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	logger.Debugf("Query: %s\nDuration: %v ms\n", event.Query, time.Since(event.StartTime).Milliseconds())
}
