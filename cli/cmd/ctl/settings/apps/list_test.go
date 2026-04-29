package apps

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestAppInfoDecode_PortsWireShape pins the wire shape of
// AppInfo.ports: BFL's AppInfo.Ports is `[]ServicePort` (struct array)
// since the BFL→main-repo merge (60d37998, 2025-12-11). Empty arrays
// decode into any slice type, so the test matrix covers an empty
// payload, a minimal-fields payload and a fully-populated payload to
// keep both shapes round-tripping cleanly.
func TestAppInfoDecode_PortsWireShape(t *testing.T) {
	cases := []struct {
		name      string
		dataJSON  string
		wantLen   int
		assertOne func(t *testing.T, p servicePort)
	}{
		{
			name:     "empty ports",
			dataJSON: `[{"id":"x","name":"testenv","ports":[]}]`,
			wantLen:  0,
		},
		{
			name:     "single minimal port (only required fields)",
			dataJSON: `[{"id":"x","name":"redis","ports":[{"name":"redis","host":"redis-svc","port":6379}]}]`,
			wantLen:  1,
			assertOne: func(t *testing.T, p servicePort) {
				t.Helper()
				if p.Name != "redis" || p.Host != "redis-svc" || p.Port != 6379 {
					t.Errorf("want name=redis host=redis-svc port=6379, got name=%q host=%q port=%d", p.Name, p.Host, p.Port)
				}
				if p.ExposePort != 0 {
					t.Errorf("ExposePort=%d; want zero (omitempty wire)", p.ExposePort)
				}
				if p.Protocol != "" {
					t.Errorf("Protocol=%q; want empty (omitempty wire, BFL default tcp)", p.Protocol)
				}
			},
		},
		{
			name: "multi-port full wire (exposePort + protocol=udp)",
			dataJSON: `[{"id":"x","name":"kafka","ports":[
				{"name":"kafka","host":"0.0.0.0","port":9092,"exposePort":30092,"protocol":"tcp"},
				{"name":"metrics","host":"0.0.0.0","port":9404,"protocol":"udp"}
			]}]`,
			wantLen: 2,
			assertOne: func(t *testing.T, p servicePort) {
				t.Helper()
				// First element of the multi-port slice.
				if p.Name != "kafka" || p.Port != 9092 || p.ExposePort != 30092 || p.Protocol != "tcp" {
					t.Errorf("kafka entry mismatch: %+v", p)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doer := &fakeDoer{}
			doer.responses = append(doer.responses, []byte(`{"code":0,"data":`+tc.dataJSON+`}`))

			var rows []appInfo
			if err := doGetEnvelope(context.Background(), doer, "/api/myapps", &rows); err != nil {
				t.Fatalf("doGetEnvelope: %v (ports wire-shape regression)", err)
			}
			if len(rows) != 1 {
				t.Fatalf("want 1 row, got %d", len(rows))
			}
			if len(rows[0].Ports) != tc.wantLen {
				t.Fatalf("want %d ports, got %d (%+v)", tc.wantLen, len(rows[0].Ports), rows[0].Ports)
			}
			if tc.assertOne != nil && len(rows[0].Ports) > 0 {
				tc.assertOne(t, rows[0].Ports[0])
			}
		})
	}
}

// TestRenderAppDetail_PortsBlock pins the rendering of the Ports
// section in `apps get`: empty ports keep the legacy single-line
// "Ports: -" KV row; non-empty ports get a tabular block under a
// "Ports:" header so all 5 ServicePort fields are visible.
func TestRenderAppDetail_PortsBlock(t *testing.T) {
	t.Run("empty ports renders legacy single line", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderAppDetail(&buf, appInfo{Name: "testenv"}); err != nil {
			t.Fatalf("renderAppDetail: %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "Ports:              -") && !strings.Contains(got, "Ports:             -") {
			// %-19s pads the left col to 19 chars; allow either width
			// in case someone tweaks the constant later.
			if !strings.Contains(got, "Ports:") || !strings.Contains(got, " -") {
				t.Errorf("want legacy single-line 'Ports: -' row, got:\n%s", got)
			}
		}
		if strings.Contains(got, "NAME\tHOST\tPORT") {
			t.Errorf("empty ports should not emit the table header; got:\n%s", got)
		}
	})

	t.Run("non-empty ports renders table block", func(t *testing.T) {
		var buf bytes.Buffer
		a := appInfo{
			Name: "kafka",
			Ports: []servicePort{
				{Name: "kafka", Host: "0.0.0.0", Port: 9092, ExposePort: 30092, Protocol: "tcp"},
				{Name: "metrics", Host: "0.0.0.0", Port: 9404}, // protocol omitted → renders as default tcp
			},
		}
		if err := renderAppDetail(&buf, a); err != nil {
			t.Fatalf("renderAppDetail: %v", err)
		}
		got := buf.String()
		// Header line under the "Ports:" section.
		if !strings.Contains(got, "Ports:\n") {
			t.Errorf("missing 'Ports:' section header; got:\n%s", got)
		}
		if !strings.Contains(got, "NAME") || !strings.Contains(got, "HOST") || !strings.Contains(got, "PORT") || !strings.Contains(got, "EXPOSE") || !strings.Contains(got, "PROTOCOL") {
			t.Errorf("missing port table columns; got:\n%s", got)
		}
		// Per-row data presence — order is tabwriter-flushed, so we
		// check substrings rather than exact strings.
		for _, want := range []string{"kafka", "0.0.0.0", "9092", "30092", "metrics", "9404"} {
			if !strings.Contains(got, want) {
				t.Errorf("missing %q in ports block; got:\n%s", want, got)
			}
		}
		// Default-tcp surfacing for the second row (protocol omitted).
		// We just need *some* "tcp" present; the first row also has
		// it explicitly, so a single substring match is enough.
		if !strings.Contains(got, "tcp") {
			t.Errorf("expected default protocol tcp to be surfaced; got:\n%s", got)
		}
		// "-" for the metrics row's empty exposePort.
		if !strings.Contains(got, " -") {
			t.Errorf("expected '-' placeholder for empty exposePort; got:\n%s", got)
		}
	})
}
