package terminus

import (
	"encoding/json"
	"testing"
)

func TestRewriteNodeScopedID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		oldNode string
		newNode string
		want    string
	}{
		{
			name:    "non-HAMI device id embeds node name",
			id:      "olares-rn5-intel-0",
			oldNode: "olares-rn5",
			newNode: "olares-rn6",
			want:    "olares-rn6-intel-0",
		},
		{
			name:    "HAMI gpu uuid is left unchanged",
			id:      "GPU-be013ee7-f327-cd75-8020-e98e278cda45",
			oldNode: "olares-rn5",
			newNode: "olares-rn6",
			want:    "GPU-be013ee7-f327-cd75-8020-e98e278cda45",
		},
		{
			name:    "unrelated id left unchanged",
			id:      "some-other-device",
			oldNode: "olares-rn5",
			newNode: "olares-rn6",
			want:    "some-other-device",
		},
		{
			name:    "empty node names are a no-op",
			id:      "olares-rn5-intel-0",
			oldNode: "",
			newNode: "olares-rn6",
			want:    "olares-rn5-intel-0",
		},
		{
			name:    "node name as exact prefix but not a segment is not rewritten",
			id:      "olares-rn5extra-intel-0",
			oldNode: "olares-rn5",
			newNode: "olares-rn6",
			// "olares-rn5extra-..." does not start with "olares-rn5-", so unchanged
			want: "olares-rn5extra-intel-0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteNodeScopedID(tt.id, tt.oldNode, tt.newNode); got != tt.want {
				t.Fatalf("rewriteNodeScopedID(%q,%q,%q) = %q, want %q", tt.id, tt.oldNode, tt.newNode, got, tt.want)
			}
		})
	}
}

func TestRewriteStringField(t *testing.T) {
	row := map[string]json.RawMessage{
		"nodeName": json.RawMessage(`"olares-rn5"`),
		"deviceId": json.RawMessage(`"olares-rn5-intel-0"`),
		"memory":   json.RawMessage(`123`),
	}

	got, changed := rewriteStringField(row, "nodeName", func(v string) string {
		if v == "olares-rn5" {
			return "olares-rn6"
		}
		return v
	})
	if !changed || got != "olares-rn6" {
		t.Fatalf("nodeName rewrite = (%q,%v), want (olares-rn6,true)", got, changed)
	}
	if string(row["nodeName"]) != `"olares-rn6"` {
		t.Fatalf("nodeName not written back: %s", row["nodeName"])
	}

	// no-op rewrite reports unchanged
	if _, changed := rewriteStringField(row, "nodeName", func(v string) string { return v }); changed {
		t.Fatalf("expected no change for identity rewrite")
	}

	// missing field is a no-op
	if _, changed := rewriteStringField(row, "absent", func(v string) string { return "x" }); changed {
		t.Fatalf("expected no change for missing field")
	}

	// non-string field is left alone
	if _, changed := rewriteStringField(row, "memory", func(v string) string { return "x" }); changed {
		t.Fatalf("expected no change for non-string field")
	}
	if string(row["memory"]) != `123` {
		t.Fatalf("memory should be untouched, got %s", row["memory"])
	}
}

func TestMigrateGPUAllocationRows(t *testing.T) {
	rows := []map[string]json.RawMessage{
		// old node, NVIDIA (stable UUID device id)
		{"nodeName": json.RawMessage(`"olares-rn7"`), "deviceId": json.RawMessage(`"GPU-abc-123"`), "owner": json.RawMessage(`"alice"`)},
		// old node, non-HAMI (node-scoped device id)
		{"nodeName": json.RawMessage(`"olares-rn7"`), "deviceId": json.RawMessage(`"olares-rn7-intel-0"`)},
		// a DIFFERENT node whose name shares the "olares-rn7-" prefix: must be untouched
		{"nodeName": json.RawMessage(`"olares-rn7-worker"`), "deviceId": json.RawMessage(`"olares-rn7-worker-intel-0"`)},
	}

	changed := migrateGPUAllocationRows(rows, "olares-rn7", "olares-rn8")
	if !changed {
		t.Fatalf("expected changes")
	}

	get := func(i int, field string) string {
		v, _ := stringField(rows[i], field)
		return v
	}
	if get(0, "nodeName") != "olares-rn8" || get(0, "deviceId") != "GPU-abc-123" {
		t.Fatalf("nvidia row = (%q,%q), want (olares-rn8, GPU-abc-123)", get(0, "nodeName"), get(0, "deviceId"))
	}
	if get(0, "owner") != "alice" {
		t.Fatalf("unmodeled field must be preserved, got owner=%q", get(0, "owner"))
	}
	if get(1, "nodeName") != "olares-rn8" || get(1, "deviceId") != "olares-rn8-intel-0" {
		t.Fatalf("intel row = (%q,%q), want (olares-rn8, olares-rn8-intel-0)", get(1, "nodeName"), get(1, "deviceId"))
	}
	if get(2, "nodeName") != "olares-rn7-worker" || get(2, "deviceId") != "olares-rn7-worker-intel-0" {
		t.Fatalf("other-node row must be untouched, got (%q,%q)", get(2, "nodeName"), get(2, "deviceId"))
	}

	// idempotent: a second run against the (already migrated) rows is a no-op
	if migrateGPUAllocationRows(rows, "olares-rn7", "olares-rn8") {
		t.Fatalf("second migration should be a no-op")
	}
}

func TestLabelAndAnnotationAllowlist(t *testing.T) {
	if !labelIsAllowlistedForRestore("gpu.bytetrade.io/nvidia") {
		t.Fatalf("gpu.bytetrade.io/* label should be allowlisted")
	}
	if !labelIsAllowlistedForRestore("node-role.kubernetes.io/worker") {
		t.Fatalf("worker role label should be allowlisted")
	}
	if labelIsAllowlistedForRestore("kubernetes.io/hostname") {
		t.Fatalf("kubernetes.io/hostname must not be restored")
	}
	if !annotationIsAllowlistedForRestore(gpuShareModeAnnotationPrefix + "GPU-abc") {
		t.Fatalf("share-mode annotation should be allowlisted")
	}
	if annotationIsAllowlistedForRestore("projectcalico.org/IPv4Address") {
		t.Fatalf("calico annotation must not be restored")
	}
}
