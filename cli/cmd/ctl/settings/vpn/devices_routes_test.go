package vpn

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestDevicesRoutesDecode_KI20 pins KI-20: headscale's `/headscale/
// machine/<id>/routes` returns route ids as strings (e.g. `"id":"3"`),
// but the CLI shipped with `route.ID int`. The bug stayed dormant
// while every probed device had no advertised routes (empty arrays
// decode regardless of element type) and surfaced on phase14b /
// phase14c the moment a populated routes payload hit the wire:
// `cannot unmarshal string into Go struct field route.routes.id of
// type int`. The fix flips ID to string and these cases lock it in.
func TestDevicesRoutesDecode_KI20(t *testing.T) {
	cases := []struct {
		name     string
		dataJSON string
		wantIDs  []string
		wantLen  int
	}{
		{
			name:     "empty routes (phase1-13/phase15a quiet shape)",
			dataJSON: `{"routes":[]}`,
			wantLen:  0,
		},
		{
			name:     "single route, string id (phase14b/c wire)",
			dataJSON: `{"routes":[{"id":"3","prefix":"100.64.0.5/32","advertised":true,"enabled":true,"isPrimary":true}]}`,
			wantIDs:  []string{"3"},
			wantLen:  1,
		},
		{
			name: "multi-route mixed enabled state",
			dataJSON: `{"routes":[
				{"id":"7","prefix":"10.0.0.0/24","advertised":true,"enabled":false,"isPrimary":false},
				{"id":"8","prefix":"10.0.1.0/24","advertised":true,"enabled":true,"isPrimary":true}
			]}`,
			wantIDs: []string{"7", "8"},
			wantLen: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp devicesRoutesResp
			if err := json.Unmarshal([]byte(tc.dataJSON), &resp); err != nil {
				t.Fatalf("unmarshal: %v (KI-20 regression: route.ID must be string, not int)", err)
			}
			if len(resp.Routes) != tc.wantLen {
				t.Fatalf("want %d routes, got %d", tc.wantLen, len(resp.Routes))
			}
			for i, want := range tc.wantIDs {
				if resp.Routes[i].ID != want {
					t.Errorf("routes[%d].ID = %q, want %q", i, resp.Routes[i].ID, want)
				}
			}
		})
	}
}

// TestDevicesRoutesDecode_RejectsLegacyIntID guards against an
// accidental rollback to `int`: feeding a numeric id on the wire
// should still decode (Go quietly handles JSON number → string when
// the target is string only via a custom unmarshaler — which we don't
// have, so it WILL error). We assert the error path stays loud so
// any future regression is obvious.
func TestDevicesRoutesDecode_RejectsLegacyIntID(t *testing.T) {
	data := `{"routes":[{"id":3,"prefix":"100.64.0.5/32"}]}`
	var resp devicesRoutesResp
	err := json.Unmarshal([]byte(data), &resp)
	if err == nil {
		t.Fatalf("want decode error for numeric id (headscale always returns string id post-KI-20); got nil — if upstream now also serves int id we should switch to json.RawMessage")
	}
	// Make sure the error is the expected type-mismatch rather than
	// some unrelated parse failure (helps next reader diagnose).
	if !strings.Contains(err.Error(), "cannot unmarshal number into Go struct field") {
		t.Errorf("unexpected error shape: %v", err)
	}
}

// TestRenderRoutesTable_StringID verifies the table renderer prints
// the new string id verbatim and does not silently coerce empties.
func TestRenderRoutesTable_StringID(t *testing.T) {
	t.Run("non-empty rows", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderRoutesTable(&buf, []route{
			{ID: "3", Prefix: "100.64.0.5/32", Advertised: true, Enabled: true, IsPrimary: true},
			{ID: "7", Prefix: "10.0.0.0/24", Advertised: true, Enabled: false},
		}); err != nil {
			t.Fatalf("render: %v", err)
		}
		got := buf.String()
		for _, want := range []string{"ID", "PREFIX", "100.64.0.5/32", "10.0.0.0/24"} {
			if !strings.Contains(got, want) {
				t.Errorf("missing %q in:\n%s", want, got)
			}
		}
		// Both ids must appear at column-0 of their rows.
		for _, want := range []string{"\n3 ", "\n7 ", "3\t", "7\t"} {
			// at-least-one-of: tabwriter padding may collapse the
			// trailing tab into spaces depending on column widths.
			if strings.Contains(got, want) {
				return
			}
		}
		t.Errorf("string ids 3/7 not visible as their own column; got:\n%s", got)
	})

	t.Run("empty id (defensive)", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderRoutesTable(&buf, []route{
			{ID: "", Prefix: "10.0.0.0/24"},
		}); err != nil {
			t.Fatalf("render: %v", err)
		}
		if !strings.Contains(buf.String(), "-") {
			t.Errorf("empty id should render as '-' via nonEmpty; got:\n%s", buf.String())
		}
	})

	t.Run("no routes at all", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderRoutesTable(&buf, nil); err != nil {
			t.Fatalf("render: %v", err)
		}
		if !strings.Contains(buf.String(), "no routes advertised") {
			t.Errorf("want 'no routes advertised' for empty list; got:\n%s", buf.String())
		}
	})
}
