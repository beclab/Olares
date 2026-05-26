package authz

import (
	"log/slog"
	"strings"
	"time"
)

// Auditor emits the 14-field ext_authz audit contract (WI-15).
type Auditor struct {
	Logger *slog.Logger
}

func (a *Auditor) logger() *slog.Logger {
	if a != nil && a.Logger != nil {
		return a.Logger
	}
	return slog.Default()
}

// Event carries one ext_authz decision audit record.
type Event struct {
	RID                 string
	Authority           string
	Method              string
	Path                string
	Decision            string
	Code                string
	Via                 string
	ViewerHost          string
	ViewerAuthenticated string
	L5dPresent          bool
	ClusterAppRef       string
	Phase               string
	LatencyMS           int64
}

// Allow logs an allow decision.
func (a *Auditor) Allow(ev Event) {
	a.emit(ev, false)
}

// Deny logs a deny decision.
func (a *Auditor) Deny(ev Event) {
	a.emit(ev, true)
}

func (a *Auditor) emit(ev Event, deny bool) {
	if ev.Decision == "" {
		if deny {
			ev.Decision = "deny"
		} else {
			ev.Decision = "allow"
		}
	}
	args := []any{
		"ts", time.Now().UTC().Format(time.RFC3339Nano),
		"rid", ev.RID,
		"authority", ev.Authority,
		"method", ev.Method,
		"path", ev.Path,
		"decision", ev.Decision,
		"code", strings.TrimSpace(ev.Code),
		"via", ev.Via,
		"viewer_host", ev.ViewerHost,
		"viewer_authenticated", ev.ViewerAuthenticated,
		"l5d_present", ev.L5dPresent,
		"cluster_app_ref", ev.ClusterAppRef,
		"phase", ev.Phase,
		"latency_ms", ev.LatencyMS,
	}
	if deny {
		a.logger().Warn("ext_authz_audit", args...)
		return
	}
	a.logger().Info("ext_authz_audit", args...)
}

func l5dPresent(headers map[string]string) bool {
	return strings.TrimSpace(headerValue(headers, "l5d-client-id")) != ""
}
