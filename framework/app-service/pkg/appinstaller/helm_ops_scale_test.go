package appinstaller

import (
	"reflect"
	"sort"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

// buildWorkloadsValues is the single source of truth for what
// .Values.workloads.* looks like in both legs of the two-phase install:
//
//   - SetValues calls it with override=0 on every install/upgrade so the
//     chart renders pods at replicas=0 regardless of the manifest.
//   - HelmOps.Scale calls it with override=replicas; in particular
//     Scale(-1) means "back to manifest-declared values" and Scale(0)
//     means "scale every workload to zero".
//
// Each of those semantics is asymmetric (override>=0 wins over the
// manifest, override<0 uses each workload's manifest-declared value;
// the manifest is required to list every workload) so it gets explicit
// truth-table coverage here. End-to-end Scale and
// SetValues invocations need a real helm action.Configuration and are
// covered by integration tests; this file pins only the deterministic
// values rendering.
func TestBuildWorkloadsValues(t *testing.T) {
	multi := appcfg.WorkloadReplicas{
		"affine":          3,
		"affine-worker":   2,
		"affine-stateful": 1,
	}
	singleZero := appcfg.WorkloadReplicas{"only": 0}

	cases := []struct {
		name     string
		wr       *appcfg.WorkloadReplicas
		override int32
		want     map[string]int32 // flattened {workloadName: replicaCount}
	}{
		{
			name:     "nil pointer renders empty map",
			wr:       nil,
			override: 0,
			want:     map[string]int32{},
		},
		{
			name:     "override=0 forces every workload to zero (SetValues path)",
			wr:       &multi,
			override: 0,
			want: map[string]int32{
				"affine":          0,
				"affine-worker":   0,
				"affine-stateful": 0,
			},
		},
		{
			name:     "override=-1 restores manifest declared values (Scale(-1) path)",
			wr:       &multi,
			override: -1,
			want: map[string]int32{
				"affine":          3,
				"affine-worker":   2,
				"affine-stateful": 1,
			},
		},
		{
			name:     "explicit-zero manifest value is preserved on restore",
			wr:       &singleZero,
			override: -1,
			want: map[string]int32{
				"only": 0,
			},
		},
		{
			name:     "explicit override value beats declared (Scale(N) path)",
			wr:       &multi,
			override: 5,
			want: map[string]int32{
				"affine":          5,
				"affine-worker":   5,
				"affine-stateful": 5,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildWorkloadsValues(tc.wr, tc.override)
			gotFlat := flattenWorkloadValues(t, got)
			if !reflect.DeepEqual(gotFlat, tc.want) {
				// Stable, diffable error: sort keys.
				t.Fatalf("rendered workloads mismatch\n got: %v\nwant: %v",
					sortedKVs(gotFlat), sortedKVs(tc.want))
			}
		})
	}
}

// TestBuildWorkloadsValuesShape confirms the sub-structure that chart
// templates rely on: each workload maps to a `replicaCount` key. Helm
// values are untyped JSON, so this is what guarantees
// `{{ .Values.workloads.<name>.replicaCount }}` in chart templates
// keeps working after refactors.
func TestBuildWorkloadsValuesShape(t *testing.T) {
	wr := appcfg.WorkloadReplicas{"affine": 7}
	out := buildWorkloadsValues(&wr, -1)

	inner, ok := out["affine"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{} for workload entry, got %T", out["affine"])
	}
	got, ok := inner["replicaCount"]
	if !ok {
		t.Fatalf("expected replicaCount key inside workload entry, got keys: %v", mapKeys(inner))
	}
	if got != int32(7) {
		t.Fatalf("replicaCount=%v (type %T), want int32(7)", got, got)
	}
}

// flattenWorkloadValues unwraps the nested map[string]interface{} layer
// buildWorkloadsValues emits so tests can use a flat want map.
func flattenWorkloadValues(t *testing.T, in map[string]interface{}) map[string]int32 {
	t.Helper()
	out := make(map[string]int32, len(in))
	for k, v := range in {
		entry, ok := v.(map[string]interface{})
		if !ok {
			t.Fatalf("workload %q entry is not a map: %T", k, v)
		}
		rc, ok := entry["replicaCount"]
		if !ok {
			t.Fatalf("workload %q entry missing replicaCount: %v", k, entry)
		}
		n, ok := rc.(int32)
		if !ok {
			t.Fatalf("workload %q replicaCount is not int32: %T", k, rc)
		}
		out[k] = n
	}
	return out
}

func sortedKVs(m map[string]int32) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+itoa(int(m[k])))
	}
	return out
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// tiny itoa to keep this file dependency-free.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
