package authz

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestAuditor_emitsRequiredFields(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	a := &Auditor{Logger: slog.New(h)}
	a.Allow(Event{
		RID: "r1", Authority: "h.example", Method: "GET", Path: "/",
		Via: "incluster_shared_allow", ViewerHost: "alice", ViewerAuthenticated: "alice",
		L5dPresent: true, Phase: "incluster", LatencyMS: 3,
	})
	out := buf.String()
	for _, key := range []string{
		"ts", "rid", "authority", "method", "path", "decision", "code", "via",
		"viewer_host", "viewer_authenticated", "l5d_present", "cluster_app_ref", "phase", "latency_ms",
	} {
		if !strings.Contains(out, key) {
			t.Fatalf("missing field %q in %s", key, out)
		}
	}
}
