package olaresclient

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
)

// fakeDoer records the last request a client shaped so tests can assert on the
// method / path / body without any network.
type fakeDoer struct {
	method string
	path   string
	body   map[string]any
}

func (d *fakeDoer) Do(_ context.Context, method, path string, body any) (json.RawMessage, error) {
	d.method = method
	d.path = path
	// Round-trip through JSON so we observe exactly what would go on the
	// wire (json tags, omitted empties), mirroring MarketClient.doRequest.
	raw, _ := json.Marshal(body)
	m := map[string]any{}
	_ = json.Unmarshal(raw, &m)
	d.body = m
	return json.RawMessage(`{}`), nil
}

func mustParse(t *testing.T, s string) *semver.Version {
	t.Helper()
	v, err := semver.NewVersion(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func TestGetClientFloorSelection(t *testing.T) {
	cases := []struct {
		name    string
		backend string
		want    string // expected Version().String() of the selected client's *registered* line
	}{
		{"exact 1.12.5", "1.12.5", "1.12.5"},
		{"exact 1.12.6", "1.12.6", "1.12.6"},
		{"1.12.5 daily", "1.12.5-20260101", "1.12.5"},
		{"1.12.6 daily", "1.12.6-20260603", "1.12.6"},
		{"1.12.6 alpha", "1.12.6-alpha1", "1.12.6"},
		{"newer than known floors to 1.12.6", "1.12.7-20260524", "1.12.6"},
		{"much newer floors to 1.12.6", "1.13.0", "1.12.6"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			backend := mustParse(t, c.backend)
			client, err := GetClient(backend)
			if err != nil {
				t.Fatalf("GetClient(%s): %v", c.backend, err)
			}
			// The client carries the FULL backend version, while the
			// chosen implementation is identified by its wire behavior.
			// Assert behavior: stop body field name distinguishes 1.12.5
			// (appName) from 1.12.6 (app_name).
			d := &fakeDoer{}
			if _, err := client.StopApp(context.Background(), d, "firefox", "market.olares", false); err != nil {
				t.Fatalf("StopApp: %v", err)
			}
			gotLine := "1.12.5"
			if _, ok := d.body["app_name"]; ok {
				gotLine = "1.12.6"
			}
			if gotLine != c.want {
				t.Fatalf("backend %s selected %s line (stop body=%v), want %s", c.backend, gotLine, d.body, c.want)
			}
			// Version() must reflect the full backend version, not the core.
			if !client.Version().Equal(backend) {
				t.Fatalf("Version()=%s, want full backend %s", client.Version(), backend)
			}
		})
	}
}

func TestGetClientNilFallsBackToDefault(t *testing.T) {
	client, err := GetClient(nil)
	if err != nil {
		t.Fatalf("GetClient(nil): %v", err)
	}
	// default behaves like 1.12.5 (appName).
	d := &fakeDoer{}
	if _, err := client.StopApp(context.Background(), d, "firefox", "market.olares", false); err != nil {
		t.Fatalf("StopApp: %v", err)
	}
	if _, ok := d.body["appName"]; !ok {
		t.Fatalf("default client should use 1.12.5 wire (appName); got body=%v", d.body)
	}
}

func TestStopResumeWireShapes(t *testing.T) {
	v5, _ := newClientV1_12_5(mustParse(t, "1.12.5"))
	v6, _ := newClientV1_12_6(mustParse(t, "1.12.6"))

	// stop
	d5 := &fakeDoer{}
	_, _ = v5.StopApp(context.Background(), d5, "firefox", "market.olares", true)
	if d5.path != "/apps/stop" || d5.body["appName"] != "firefox" {
		t.Fatalf("1.12.5 stop: path=%s body=%v", d5.path, d5.body)
	}
	if _, ok := d5.body["app_name"]; ok {
		t.Fatalf("1.12.5 stop must NOT carry app_name: %v", d5.body)
	}
	if _, ok := d5.body["source"]; ok {
		t.Fatalf("1.12.5 stop must NOT carry source: %v", d5.body)
	}

	d6 := &fakeDoer{}
	_, _ = v6.StopApp(context.Background(), d6, "firefox", "market.olares", true)
	if d6.path != "/apps/stop" || d6.body["app_name"] != "firefox" || d6.body["source"] != "market.olares" {
		t.Fatalf("1.12.6 stop must carry app_name+source: path=%s body=%v", d6.path, d6.body)
	}
	if _, ok := d6.body["appName"]; ok {
		t.Fatalf("1.12.6 stop must NOT carry appName: %v", d6.body)
	}

	// resume
	r5 := &fakeDoer{}
	_, _ = v5.ResumeApp(context.Background(), r5, "ollama", "market.olares")
	if r5.body["appName"] != "ollama" {
		t.Fatalf("1.12.5 resume body=%v", r5.body)
	}
	if _, ok := r5.body["source"]; ok {
		t.Fatalf("1.12.5 resume must NOT carry source: %v", r5.body)
	}
	r6 := &fakeDoer{}
	_, _ = v6.ResumeApp(context.Background(), r6, "ollama", "market.olares")
	if r6.body["app_name"] != "ollama" || r6.body["source"] != "market.olares" {
		t.Fatalf("1.12.6 resume must carry app_name+source: %v", r6.body)
	}
}

func TestUninstallWireShapes(t *testing.T) {
	v5, _ := newClientV1_12_5(mustParse(t, "1.12.5"))
	v6, _ := newClientV1_12_6(mustParse(t, "1.12.6"))

	d5 := &fakeDoer{}
	_, _ = v5.UninstallApp(context.Background(), d5, "windows", "market.olares", "1.2.3", true, true)
	if d5.method != "DELETE" || d5.path != "/apps/windows" {
		t.Fatalf("1.12.5 uninstall method=%s path=%s", d5.method, d5.path)
	}
	// 1.12.5 does NOT send app_name / source / version in the body.
	for _, k := range []string{"version", "app_name", "source"} {
		if _, ok := d5.body[k]; ok {
			t.Fatalf("1.12.5 uninstall must NOT carry %s: %v", k, d5.body)
		}
	}

	d6 := &fakeDoer{}
	_, _ = v6.UninstallApp(context.Background(), d6, "windows", "market.olares", "1.2.3", true, true)
	if d6.body["app_name"] != "windows" || d6.body["source"] != "market.olares" || d6.body["version"] != "1.2.3" {
		t.Fatalf("1.12.6 uninstall must carry app_name+source+version: %v", d6.body)
	}

	// version omitted when empty even on 1.12.6 (source still required).
	d6e := &fakeDoer{}
	_, _ = v6.UninstallApp(context.Background(), d6e, "windows", "market.olares", "", true, true)
	if _, ok := d6e.body["version"]; ok {
		t.Fatalf("1.12.6 uninstall should omit empty version: %v", d6e.body)
	}
	if d6e.body["source"] != "market.olares" {
		t.Fatalf("1.12.6 uninstall must still carry source: %v", d6e.body)
	}
}

func TestMethodFallbackThroughEmbedding(t *testing.T) {
	// clientV1_12_6 embeds clientV1_12_5; a *V1_12_6 must still satisfy the
	// full interface, and Version() (defined on baseClient, reused via
	// embedding) must work without an explicit override.
	v6, err := newClientV1_12_6(mustParse(t, "1.12.6-20260603"))
	if err != nil {
		t.Fatalf("new 1.12.6: %v", err)
	}
	if got := v6.Version().String(); got != "1.12.6-20260603" {
		t.Fatalf("Version()=%s, want 1.12.6-20260603 (inherited from baseClient)", got)
	}
}

// --- ComputeOps ---

func TestComputeWireShapes_1_12_6(t *testing.T) {
	v6, _ := newClientV1_12_6(mustParse(t, "1.12.6"))

	d := &fakeDoer{}
	if _, err := v6.ListAccelerators(context.Background(), d); err != nil {
		t.Fatalf("ListAccelerators: %v", err)
	}
	if d.method != "GET" || d.path != "/api/compute-resources" {
		t.Fatalf("1.12.6 list: method=%s path=%s", d.method, d.path)
	}

	gb := &fakeDoer{}
	if _, err := v6.GetAppBindings(context.Background(), gb, "ollama"); err != nil {
		t.Fatalf("GetAppBindings: %v", err)
	}
	if gb.method != "GET" || gb.path != "/api/apps/ollama/compute-resources/bindings" {
		t.Fatalf("1.12.6 bindings: method=%s path=%s", gb.method, gb.path)
	}

	rl := &fakeDoer{}
	if _, err := v6.ReleaseAppBindings(context.Background(), rl, "ollama"); err != nil {
		t.Fatalf("ReleaseAppBindings: %v", err)
	}
	if rl.method != "DELETE" || rl.path != "/api/apps/ollama/compute-resources/bindings" {
		t.Fatalf("1.12.6 unbind: method=%s path=%s", rl.method, rl.path)
	}

	sw := &fakeDoer{}
	if _, err := v6.SwitchSupportType(context.Background(), sw, "node-1", "GPU-abc", "TimeSlice"); err != nil {
		t.Fatalf("SwitchSupportType: %v", err)
	}
	if sw.method != "PUT" || sw.path != "/api/compute-resources/nodes/node-1/devices/GPU-abc/support-type" {
		t.Fatalf("1.12.6 support-type: method=%s path=%s", sw.method, sw.path)
	}
	if sw.body["supportType"] != "TimeSlice" {
		t.Fatalf("1.12.6 support-type body=%v", sw.body)
	}
}

func TestComputeListLegacy_1_12_5(t *testing.T) {
	v5, _ := newClientV1_12_5(mustParse(t, "1.12.5"))
	d := &fakeDoer{}
	if _, err := v5.ListAccelerators(context.Background(), d); err != nil {
		t.Fatalf("ListAccelerators: %v", err)
	}
	if d.method != "GET" || d.path != "/api/gpu/list" {
		t.Fatalf("1.12.5 list must hit legacy endpoint: method=%s path=%s", d.method, d.path)
	}
}

func TestComputeGatingOnLegacy_1_12_5(t *testing.T) {
	v5, _ := newClientV1_12_5(mustParse(t, "1.12.5"))
	d := &fakeDoer{}

	cases := []struct {
		name string
		call func() error
	}{
		{"bindings", func() error { _, err := v5.GetAppBindings(context.Background(), d, "ollama"); return err }},
		{"unbind", func() error { _, err := v5.ReleaseAppBindings(context.Background(), d, "ollama"); return err }},
		{"support-type", func() error {
			_, err := v5.SwitchSupportType(context.Background(), d, "n", "dev", "Exclusive")
			return err
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.call()
			var unsupported *ErrUnsupportedVersion
			if !errors.As(err, &unsupported) {
				t.Fatalf("%s on 1.12.5 must return ErrUnsupportedVersion, got %v", c.name, err)
			}
			if unsupported.MinVersion == nil || unsupported.MinVersion.String() != "1.12.6" {
				t.Fatalf("%s gate must require 1.12.6, got %v", c.name, unsupported.MinVersion)
			}
		})
	}
}

// --- OverlayOps ---

func TestOverlayWireShapes_1_12_6(t *testing.T) {
	v6, _ := newClientV1_12_6(mustParse(t, "1.12.6"))

	st := &fakeDoer{}
	if _, err := v6.OverlayGatewayStatus(context.Background(), st, "alice"); err != nil {
		t.Fatalf("OverlayGatewayStatus: %v", err)
	}
	if st.method != "GET" || st.path != "/api/system/overlay-gateway-status/alice" {
		t.Fatalf("1.12.6 overlay status: method=%s path=%s", st.method, st.path)
	}

	en := &fakeDoer{}
	if _, err := v6.EnableOverlayGateway(context.Background(), en); err != nil {
		t.Fatalf("EnableOverlayGateway: %v", err)
	}
	if en.method != "POST" || en.path != "/api/command/enable-overlay-gateway" {
		t.Fatalf("1.12.6 overlay enable: method=%s path=%s", en.method, en.path)
	}

	dis := &fakeDoer{}
	if _, err := v6.DisableOverlayGateway(context.Background(), dis); err != nil {
		t.Fatalf("DisableOverlayGateway: %v", err)
	}
	if dis.method != "POST" || dis.path != "/api/command/disable-overlay-gateway" {
		t.Fatalf("1.12.6 overlay disable: method=%s path=%s", dis.method, dis.path)
	}
}

func TestOverlayGatingOnLegacy_1_12_5(t *testing.T) {
	v5, _ := newClientV1_12_5(mustParse(t, "1.12.5"))
	d := &fakeDoer{}

	cases := []struct {
		name string
		call func() error
	}{
		{"status", func() error { _, err := v5.OverlayGatewayStatus(context.Background(), d, "alice"); return err }},
		{"enable", func() error { _, err := v5.EnableOverlayGateway(context.Background(), d); return err }},
		{"disable", func() error { _, err := v5.DisableOverlayGateway(context.Background(), d); return err }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.call()
			var unsupported *ErrUnsupportedVersion
			if !errors.As(err, &unsupported) {
				t.Fatalf("overlay %s on 1.12.5 must return ErrUnsupportedVersion, got %v", c.name, err)
			}
		})
	}
}

// TestDefaultClientCapabilities ensures the fallback (nil version) client
// inherits the conservative 1.12.5 behavior: legacy GPU list, and the
// compute/overlay extras gated off.
func TestDefaultClientCapabilities(t *testing.T) {
	client, err := GetClient(nil)
	if err != nil {
		t.Fatalf("GetClient(nil): %v", err)
	}
	d := &fakeDoer{}
	if _, err := client.ListAccelerators(context.Background(), d); err != nil {
		t.Fatalf("default ListAccelerators: %v", err)
	}
	if d.path != "/api/gpu/list" {
		t.Fatalf("default client should use legacy gpu list, got path=%s", d.path)
	}
	if _, err := client.EnableOverlayGateway(context.Background(), d); err == nil {
		t.Fatal("default client EnableOverlayGateway must be gated")
	} else {
		var unsupported *ErrUnsupportedVersion
		if !errors.As(err, &unsupported) {
			t.Fatalf("default overlay enable must return ErrUnsupportedVersion, got %v", err)
		}
	}
}

func TestErrUnsupportedVersionMessage(t *testing.T) {
	err := &ErrUnsupportedVersion{
		Feature:    "market clone --entrance",
		MinVersion: mustParse(t, "1.12.6"),
		Current:    mustParse(t, "1.12.5"),
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("empty error message")
	}
	// Must mention both the requirement and the current version.
	for _, want := range []string{"1.12.6", "1.12.5", "market clone --entrance"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}
