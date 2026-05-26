package market

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestIsCSV2 pins isCSV2's predicate to the SPA's
// apps/.../constant/constants.ts implementation: TRUE only when both
//
//	app_info.app_entry.apiVersion == "v2"
//	&& len(app_info.app_entry.subCharts) > 0
//
// hold. Drift in either condition would silently change the auto-cascade
// default in `market uninstall` — keep this table in lockstep with the
// SPA's isCSV2() (see csAppUninstall in appService.ts).
func TestIsCSV2(t *testing.T) {
	cases := []struct {
		name   string
		input  map[string]interface{}
		expect bool
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: false,
		},
		{
			name:   "empty map",
			input:  map[string]interface{}{},
			expect: false,
		},
		{
			name: "missing app_entry",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{},
			},
			expect: false,
		},
		{
			name: "v2 with non-empty subCharts → true",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v2",
						"subCharts": []interface{}{
							map[string]interface{}{"name": "server"},
							map[string]interface{}{"name": "client"},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "v2 with empty subCharts → false (single-chart v2 app)",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v2",
						"subCharts":  []interface{}{},
					},
				},
			},
			expect: false,
		},
		{
			name: "v2 missing subCharts → false",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v2",
					},
				},
			},
			expect: false,
		},
		{
			name: "v1 with subCharts → false (apiVersion gate)",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v1",
						"subCharts": []interface{}{
							map[string]interface{}{"name": "server"},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "subCharts wrong type (object, not array) → false",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v2",
						"subCharts":  map[string]interface{}{"name": "server"},
					},
				},
			},
			expect: false,
		},
		{
			name: "apiVersion missing → false",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"subCharts": []interface{}{
							map[string]interface{}{"name": "server"},
						},
					},
				},
			},
			expect: false,
		},
		{
			// Regression guard: clusterScoped on its own must NOT
			// trigger CS — the cascade check is about v2 multi-chart,
			// not about cluster scope. SPA keeps these orthogonal too.
			name: "clusterScoped without v2 subCharts → false",
			input: map[string]interface{}{
				"app_info": map[string]interface{}{
					"app_entry": map[string]interface{}{
						"apiVersion": "v1",
						"options": map[string]interface{}{
							"appScope": map[string]interface{}{
								"clusterScoped": true,
							},
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isCSV2(c.input); got != c.expect {
				t.Fatalf("isCSV2(%s) = %v, want %v", c.name, got, c.expect)
			}
		})
	}
}

// TestShouldAutoCascadeWith locks down shouldAutoCascade's decision
// matrix WITHOUT requiring a real Factory / HTTP stack — the
// cascadeProbe seam lets tests stub out the three production calls
// (user totals, state lookup, catalog metadata). The critical case is
// the clone path: when row.RawName is set (e.g. windowsefe992 → windows),
// the catalog lookup MUST use RawName, not the per-instance app name,
// or the /apps response is empty, isCSV2 returns false, and the
// auto-cascade default silently diverges from the SPA's
// csAppUninstall() / csAppStop() — the exact bug fixed by switching
// shouldAutoCascade to lookupInstalledApp + RawName-preferred lookup.
//
// Other cases pin the surrounding branches: multi-user short-circuit
// (no CS probe), zero-user / probe-error soft-fail, missing state row
// (no error, just "not a CS bundle as far as we can tell"), non-CS-V2
// catalog metadata (apiVersion != v2), and the standard non-clone
// happy path.
func TestShouldAutoCascadeWith(t *testing.T) {
	v2MultiChart := map[string]interface{}{
		"app_info": map[string]interface{}{
			"app_entry": map[string]interface{}{
				"apiVersion": "v2",
				"subCharts": []interface{}{
					map[string]interface{}{"name": "server"},
				},
			},
		},
	}
	v1Single := map[string]interface{}{
		"app_info": map[string]interface{}{
			"app_entry": map[string]interface{}{
				"apiVersion": "v1",
			},
		},
	}

	cases := []struct {
		name    string
		appName string
		probe   cascadeProbe

		wantAuto         bool
		wantWhyContains  []string
		wantLookupName   string // the catalog key the probe MUST receive
		wantLookupSource string
	}{
		{
			name:    "clone path: RawName used for catalog, isCSV2 → auto-cascade on",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					if name != "windowsefe992" {
						t.Fatalf("lookupRow received unexpected name %q", name)
					}
					return &installedAppRow{Name: "windowsefe992", RawName: "windows", Source: "market.olares"}, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					if name != "windows" {
						t.Fatalf("clone catalog lookup hit the wrong key: got %q want %q", name, "windows")
					}
					if src != "market.olares" {
						t.Fatalf("clone catalog lookup source: got %q want %q", src, "market.olares")
					}
					return v2MultiChart, nil
				},
			},
			wantAuto:         true,
			wantWhyContains:  []string{"via source app \"windows\"", "market.olares"},
			wantLookupName:   "windows",
			wantLookupSource: "market.olares",
		},
		{
			name:    "non-clone path: Name used for catalog, isCSV2 → auto-cascade on",
			appName: "ollamav2",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					return &installedAppRow{Name: "ollamav2", RawName: "", Source: "market.olares"}, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					if name != "ollamav2" {
						t.Fatalf("non-clone catalog lookup key: got %q want %q", name, "ollamav2")
					}
					return v2MultiChart, nil
				},
			},
			wantAuto:        true,
			wantWhyContains: []string{"v2 multi-chart", "market.olares"},
		},
		{
			name:    "clone path: RawName=Name (degenerate) goes through the Name branch",
			appName: "windows",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					return &installedAppRow{Name: "windows", RawName: "windows", Source: "market.olares"}, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					if name != "windows" {
						t.Fatalf("degenerate clone catalog lookup key: got %q want %q", name, "windows")
					}
					return v2MultiChart, nil
				},
			},
			wantAuto: true,
			// Reason should NOT mention "via source app" since the
			// lookup name and app name match — keeps the stderr
			// message accurate for normal (non-clone) installs.
			wantWhyContains: []string{"v2 multi-chart"},
		},
		{
			name:    "multi-user → fast path skips CS probe",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 2, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					t.Fatal("lookupRow should not be called when totals > 1")
					return nil, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					t.Fatal("fetchAppMeta should not be called when totals > 1")
					return nil, nil
				},
			},
			wantAuto: false,
		},
		{
			name:    "user-totals probe error → soft-fail false",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals:  func(ctx context.Context) (int, error) { return 0, errors.New("synthetic /api/users/v2 error") },
				lookupRow:    func(ctx context.Context, name string) (*installedAppRow, error) { return nil, nil },
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) { return nil, nil },
			},
			wantAuto: false,
		},
		{
			name:    "totals=0 → soft-fail false (defensive zero handling)",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals:  func(ctx context.Context) (int, error) { return 0, nil },
				lookupRow:    func(ctx context.Context, name string) (*installedAppRow, error) { return nil, nil },
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) { return nil, nil },
			},
			wantAuto: false,
		},
		{
			name:    "row not found → soft-fail false (let DELETE/POST surface the 404)",
			appName: "ghost",
			probe: cascadeProbe{
				fetchTotals:  func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow:    func(ctx context.Context, name string) (*installedAppRow, error) { return nil, nil },
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) { return nil, nil },
			},
			wantAuto: false,
		},
		{
			name:    "state probe error → soft-fail false",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals:  func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow:    func(ctx context.Context, name string) (*installedAppRow, error) { return nil, errors.New("synthetic state error") },
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) { return nil, nil },
			},
			wantAuto: false,
		},
		{
			name:    "catalog probe error → soft-fail false",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					return &installedAppRow{Name: "windowsefe992", RawName: "windows", Source: "market.olares"}, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					return nil, errors.New("synthetic /apps error")
				},
			},
			wantAuto: false,
		},
		{
			name:    "not CS-V2 (v1 single-chart) → false",
			appName: "windowsefe992",
			probe: cascadeProbe{
				fetchTotals: func(ctx context.Context) (int, error) { return 1, nil },
				lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
					return &installedAppRow{Name: "windowsefe992", RawName: "windows", Source: "market.olares"}, nil
				},
				fetchAppMeta: func(ctx context.Context, name, src string) (map[string]interface{}, error) {
					if name != "windows" {
						t.Fatalf("catalog lookup must use RawName even when result is v1: got %q", name)
					}
					return v1Single, nil
				},
			},
			wantAuto: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotAuto, gotWhy := shouldAutoCascadeWith(context.Background(), c.probe, c.appName)
			if gotAuto != c.wantAuto {
				t.Fatalf("auto = %v, want %v (why=%q)", gotAuto, c.wantAuto, gotWhy)
			}
			for _, frag := range c.wantWhyContains {
				if !strings.Contains(gotWhy, frag) {
					t.Fatalf("why=%q must contain %q", gotWhy, frag)
				}
			}
			if c.wantAuto && c.wantWhyContains == nil && gotWhy == "" {
				t.Fatalf("auto-true case must populate why; got empty string")
			}
		})
	}
}
