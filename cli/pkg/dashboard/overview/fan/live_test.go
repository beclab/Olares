package fan

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestBuildLiveEnvelope_HappyPath pins the SPA-aligned single-row
// envelope: live RPM/temperature from /system/cpu-gpu joined with
// power/powerLimit from the (optional) HAMI graphics list. Display
// columns must use the SPA's exact format strings ("%.0f / %d
// RPM", "%.2f W", etc.) so agent scrapers don't break.
func TestBuildLiveEnvelope_HappyPath(t *testing.T) {
	srv := fanStubMux{
		systemStatus: olaresOneStatus,
		systemFan: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"code":0,"data":{"cpu_fan_speed":1500,"cpu_temperature":72.5,"gpu_fan_speed":2100,"gpu_temperature":68}}`))
		},
		graphics: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[{"power":35.75,"powerLimit":120}]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildLiveEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildLiveEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewFanLive {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewFanLive)
	}
	if len(env.Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(env.Items))
	}
	row := env.Items[0]
	// Raw payload mirrors the SPA's flat numeric record so JSON
	// consumers can post-format without re-parsing.
	if row.Raw["cpu_fan_rpm"] != 1500.0 {
		t.Errorf("cpu_fan_rpm = %v, want 1500", row.Raw["cpu_fan_rpm"])
	}
	if row.Raw["gpu_power"] != 35.75 || row.Raw["gpu_power_limit"] != 120.0 {
		t.Errorf("gpu_power/limit = %v / %v, want 35.75 / 120", row.Raw["gpu_power"], row.Raw["gpu_power_limit"])
	}
	// Display formatting is column-stable.
	if got := row.Display["cpu_fan"]; got != "1500 / "+itoa(pkgdashboard.FanSpeedMaxCPU)+" RPM" {
		t.Errorf("cpu_fan = %q, want \"<rpm> / <max> RPM\" exact format", got)
	}
	if got := row.Display["gpu_power"]; got != "35.75 W" {
		t.Errorf("gpu_power = %q, want \"35.75 W\"", got)
	}
	if got := row.Display["gpu_power_lim"]; got != "120 W" {
		t.Errorf("gpu_power_lim = %q, want \"120 W\"", got)
	}
}

// TestBuildLiveEnvelope_NoFanIntegration404 pins the missing-fan
// branch: a 404 from /system/cpu-gpu must surface as Empty=true +
// EmptyReason="no_fan_integration" rather than as a transport
// error, so JSON consumers can demux without inspecting err.Error().
func TestBuildLiveEnvelope_NoFanIntegration404(t *testing.T) {
	srv := fanStubMux{
		systemStatus: olaresOneStatus,
		systemFan: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildLiveEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildLiveEnvelope: unexpected err: %v", err)
	}
	if !env.Meta.Empty || env.Meta.EmptyReason != "no_fan_integration" {
		t.Errorf("Meta = %+v, want Empty=true / EmptyReason=no_fan_integration", env.Meta)
	}
	if env.Meta.HTTPStatus != http.StatusNotFound {
		t.Errorf("Meta.HTTPStatus = %d, want 404", env.Meta.HTTPStatus)
	}
}

// TestBuildLiveEnvelope_NoGraphicsIntegrationStillSucceeds pins the
// degraded path: when HAMI returns 404 the live envelope still
// emits a row, with gpu_power/limit defaulting to 0. This is the
// "fan present, no GPU" case (CPU-only Olares One).
func TestBuildLiveEnvelope_NoGraphicsIntegrationStillSucceeds(t *testing.T) {
	srv := fanStubMux{
		systemStatus: olaresOneStatus,
		systemFan: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"code":0,"data":{"cpu_fan_speed":1100,"cpu_temperature":54,"gpu_fan_speed":0,"gpu_temperature":0}}`))
		},
		graphics: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildLiveEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildLiveEnvelope: %v", err)
	}
	if len(env.Items) != 1 {
		t.Fatalf("Items len = %d, want 1 (fan present)", len(env.Items))
	}
	if env.Items[0].Raw["gpu_power"] != 0.0 || env.Items[0].Raw["gpu_power_limit"] != 0.0 {
		t.Errorf("gpu_power/limit = %v / %v, want 0 / 0 (HAMI 404 → defaults)",
			env.Items[0].Raw["gpu_power"], env.Items[0].Raw["gpu_power_limit"])
	}
	got, _ := env.Items[0].Display["gpu_power"].(string)
	if !strings.HasPrefix(got, "0.00") {
		t.Errorf("gpu_power display = %q, want \"0.00 W\"-shape", got)
	}
}

// itoa is a tiny inline helper to keep the format-assertion call
// sites readable. We avoid strconv to stay cohesive.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
