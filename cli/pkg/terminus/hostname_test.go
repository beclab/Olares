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
