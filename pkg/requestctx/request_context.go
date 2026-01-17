package requestctx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type ctxKeyType struct{}

var ctxKey = ctxKeyType{}

type RequestContext struct {
	// Auth
	Authorization   string `json:"authorization"`
	BasicAuth       string `json:"basic_auth"`
	Timezone        string `json:"timezone"`
	ValidatedClaims any    `json:"validated_claims,omitempty"`

	// Custom app headers
	UserName string `json:"user_name"`
	Version  string `json:"version"`
	Language string `json:"language"`
	Platform string `json:"platform"`

	// User info
	UserID   string `json:"user_id"`
	UserUUID string `json:"user_uuid"`

	// TODO Auth
	//ValidatedClaims jwt.Claims `json:"validated_claims,omitempty"`

	// RBAC - Multi-tenant context
	TenantID string `json:"tenant_id,omitempty"`

	// Meta
	RequestID     string            `json:"request_id,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	ClientIP      string            `json:"client_ip,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	IsGuest       bool              `json:"is_guest"`
	RawHeaders    map[string]string `json:"-"`

	CreatedAt time.Time `json:"-"`
}

func FromRequest(r *http.Request) *RequestContext {
	h := r.Header
	rc := &RequestContext{
		RawHeaders:    make(map[string]string),
		CreatedAt:     time.Now(),
		UserAgent:     r.UserAgent(),
		Authorization: h.Get("Authorization"),
		BasicAuth:     h.Get("BasicAuth"),
		Timezone:      defaultIfEmpty(h.Get("Timezone"), "+07:00"),

		UserName: defaultIfEmpty(h.Get("X-Username"), "default_noname"),
		Version:  defaultIfEmpty(h.Get("X-Version"), "default_0.0.0"),
		Language: defaultIfEmpty(h.Get("X-Lang"), "ID"),
		Platform: defaultIfEmpty(h.Get("X-Device"), "default_platform"),

		UserID:   defaultIfEmpty(h.Get("X-User-Id"), "default_id"),
		UserUUID: defaultIfEmpty(h.Get("X-User-Uuid"), "default_uuid"),

		TenantID: h.Get("X-Tenant-Id"), // Multi-tenant context

		RequestID:     h.Get("X-Request-Id"),
		CorrelationID: h.Get("X-Correlation-Id"),

		ClientIP: clientIPFromRequest(r),
		IsGuest:  true,
	}

	// Copy all raw headers (for debugging / tracing)
	for k, v := range h {
		if len(v) > 0 {
			rc.RawHeaders[k] = v[0]
		}
	}

	rc.IsGuest = rc.UserID == "" || strings.EqualFold(rc.Authorization, "guest")
	return rc
}

func defaultIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func clientIPFromRequest(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return strings.Split(xf, ",")[0]
	}
	if xr := r.Header.Get("X-Real-Ip"); xr != "" {
		return xr
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func NewContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, ctxKey, rc)
}

func FromContext(ctx context.Context) *RequestContext {
	if ctx == nil {
		return &RequestContext{IsGuest: true}
	}
	if v := ctx.Value(ctxKey); v != nil {
		if rc, ok := v.(*RequestContext); ok {
			return rc
		}
	}
	return &RequestContext{IsGuest: true}
}

func GetUserID(ctx context.Context) string {
	rc := FromContext(ctx)
	if rc == nil {
		return ""
	}
	return rc.UserID
}

// GetUserIDAsInt64 returns the user ID as int64 (for RBAC)
func GetUserIDAsInt64(ctx context.Context) int64 {
	rc := FromContext(ctx)
	if rc == nil || rc.UserID == "" {
		return 0
	}
	// Try to parse the user ID as int64
	var id int64
	_, err := fmt.Sscanf(rc.UserID, "%d", &id)
	if err != nil {
		return 0
	}
	return id
}

// GetTenantID returns the tenant ID from context
func GetTenantID(ctx context.Context) string {
	rc := FromContext(ctx)
	if rc == nil {
		return ""
	}
	return rc.TenantID
}

// SetTenantID sets the tenant ID in the request context
func SetTenantID(ctx context.Context, tenantID string) context.Context {
	rc := FromContext(ctx)
	if rc != nil {
		rc.TenantID = tenantID
	}
	return NewContext(ctx, rc)
}
