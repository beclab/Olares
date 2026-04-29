package fan

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestBuildSectionsEnvelope_HappyPath covers the standard fan-out:
// live (1 row) + curve (10 rows) co-emitted under a single
// dashboard envelope. liveErr is nil because the fan endpoint
// resolved cleanly.
func TestBuildSectionsEnvelope_HappyPath(t *testing.T) {
	srv := fanStubMux{
		systemStatus: olaresOneStatus,
		systemFan: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"code":0,"data":{"cpu_fan_speed":1500,"cpu_temperature":72,"gpu_fan_speed":2100,"gpu_temperature":68}}`))
		},
		graphics: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[{"power":35,"powerLimit":120}]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, lerr := BuildSectionsEnvelope(context.Background(), c, cf, time.Now())
	if lerr != nil {
		t.Fatalf("BuildSectionsEnvelope live err: %v", lerr)
	}
	if env.Kind != pkgdashboard.KindOverviewFan {
		t.Errorf("parent Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewFan)
	}
	live, ok := env.Sections["live"]
	if !ok {
		t.Fatal("section `live` missing")
	}
	if len(live.Items) != 1 {
		t.Errorf("live items = %d, want 1", len(live.Items))
	}
	curve, ok := env.Sections["curve"]
	if !ok {
		t.Fatal("section `curve` missing")
	}
	if got, want := len(curve.Items), len(pkgdashboard.FanCurveTable); got != want {
		t.Errorf("curve items = %d, want %d", got, want)
	}
}

// TestBuildSectionsEnvelope_LiveErrorPropagatedToMeta covers the
// degraded path: when the live fetch fails (5xx) the curve still
// emits but the live section's Meta.Error / ErrorKind populates so
// JSON consumers can detect the partial-failure mode without the
// (returned) lerr.
func TestBuildSectionsEnvelope_LiveErrorPropagatedToMeta(t *testing.T) {
	srv := fanStubMux{
		systemStatus: olaresOneStatus,
		systemFan: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, lerr := BuildSectionsEnvelope(context.Background(), c, cf, time.Now())
	if lerr == nil {
		t.Fatal("BuildSectionsEnvelope: expected liveErr, got nil")
	}
	live := env.Sections["live"]
	if live.Meta.Error == "" {
		t.Errorf("live.Meta.Error empty, want propagated 5xx string")
	}
	if live.Meta.ErrorKind == "" {
		t.Errorf("live.Meta.ErrorKind empty, want classified")
	}
	curve := env.Sections["curve"]
	if got, want := len(curve.Items), len(pkgdashboard.FanCurveTable); got != want {
		t.Errorf("curve items = %d, want %d (curve must still emit on live failure)", got, want)
	}
}

// TestRunDefault_NotOlaresOneEmitsGatedJSON covers the capability
// gate: on a non-Olares-One device the parent envelope carries
// Empty=true / EmptyReason="not_olares_one" + per-section gating
// (live + curve both empty), and JSON output emits the parent
// envelope so consumers can demux either at the top or per-section.
func TestRunDefault_NotOlaresOneEmitsGatedJSON(t *testing.T) {
	srv := fanStubMux{systemStatus: genericStatus}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	stdout := captureStdout(t, func() error {
		return RunDefault(context.Background(), c, cf)
	})
	if !strings.Contains(stdout, `"empty_reason":"not_olares_one"`) {
		t.Errorf("RunDefault stdout missing not_olares_one reason; got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"sections"`) {
		t.Errorf("RunDefault stdout missing sections envelope; got:\n%s", stdout)
	}
}

// TestWriteSectionsTable_BannersAndLiveErrorPath covers the human
// scrollback layout: "== LIVE ==" + "== CURVE ==" banners, with the
// live error replaced by an "(error: ...)" line when liveErr is
// set, and the curve still rendering after.
func TestWriteSectionsTable_BannersAndLiveErrorPath(t *testing.T) {
	cf := fixtureFlags(t)
	curveEnv := BuildCurveEnvelope(cf, "alice@olares.com", time.Now())
	env := pkgdashboard.Envelope{
		Sections: map[string]pkgdashboard.Envelope{
			"live":  {Kind: pkgdashboard.KindOverviewFanLive},
			"curve": curveEnv,
		},
	}
	var buf bytes.Buffer
	if err := WriteSectionsTable(&buf, env, errBoom("hami down")); err != nil {
		t.Fatalf("WriteSectionsTable: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "== LIVE ==") || !strings.Contains(out, "== CURVE ==") {
		t.Errorf("missing banners; out:\n%s", out)
	}
	if !strings.Contains(out, "(error: hami down)") {
		t.Errorf("missing live error line; out:\n%s", out)
	}
	if !strings.Contains(out, "STEP") {
		t.Errorf("curve table did not render; out:\n%s", out)
	}
}

// errBoom is a tiny string error used by TestWriteSectionsTable to
// inject a synthetic live error without dragging fmt.Errorf imports.
type errBoom string

func (e errBoom) Error() string { return string(e) }
