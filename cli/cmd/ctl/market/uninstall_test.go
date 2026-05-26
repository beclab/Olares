package market

import (
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
