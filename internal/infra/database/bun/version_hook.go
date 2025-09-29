package bun

import (
	"context"
	"github.com/uptrace/bun"
	"time"
)

type VersionsSetter interface {
	SetVersions(int64)
}

type VersionMixin struct{}

func (VersionMixin) BeforeAppendModel(ctx context.Context, q bun.Query) error {
	switch q.Operation() {
	case "INSERT", "UPDATE", "DELETE":
		if m, ok := q.GetModel().Value().(VersionsSetter); ok {
			m.SetVersions(time.Now().UnixNano())
		}
	}
	return nil
}
