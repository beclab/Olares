package pod

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// TestTranslatePodFieldSelector covers the kubectl-syntax →
// KubeSphere query-param mapping for cluster pod list. The
// per-case rationale is on each entry; the suite as a whole guards
// against silently-empty list responses when a user passes the
// standard kubectl selector shape the upstream KubeSphere endpoint
// doesn't natively understand.
func TestTranslatePodFieldSelector(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    url.Values
		wantErr string // substring; empty means success
	}{
		{
			name: "empty returns empty values",
			in:   "",
			want: url.Values{},
		},
		{
			name: "whitespace only returns empty values",
			in:   "   ",
			want: url.Values{},
		},
		{
			name: "status.phase=Running maps to status=Running",
			in:   "status.phase=Running",
			want: url.Values{"status": {"Running"}},
		},
		{
			name: "double-equal accepted same as single",
			in:   "status.phase==Running",
			want: url.Values{"status": {"Running"}},
		},
		{
			name: "spec.nodeName=worker-1 maps to nodeName",
			in:   "spec.nodeName=worker-1",
			want: url.Values{"nodeName": {"worker-1"}},
		},
		{
			name: "metadata.name maps to names (exact match key)",
			in:   "metadata.name=foo",
			want: url.Values{"names": {"foo"}},
		},
		{
			name: "metadata.namespace maps to namespace",
			in:   "metadata.namespace=kube-system",
			want: url.Values{"namespace": {"kube-system"}},
		},
		{
			name: "multiple terms combine",
			in:   "status.phase=Running,spec.nodeName=worker-1",
			want: url.Values{"status": {"Running"}, "nodeName": {"worker-1"}},
		},
		{
			name: "trailing comma tolerated",
			in:   "status.phase=Running,",
			want: url.Values{"status": {"Running"}},
		},
		{
			name: "extra whitespace around term and operator",
			in:   "  status.phase = Running ",
			want: url.Values{"status": {"Running"}},
		},
		{
			name:    "not-equal rejected (KubeSphere doesn't support it)",
			in:      "status.phase!=Running",
			wantErr: "'!=' operator",
		},
		{
			name:    "missing operator rejected",
			in:      "status.phase Running",
			wantErr: "not a valid term",
		},
		{
			name:    "empty value rejected",
			in:      "status.phase=",
			wantErr: "empty key or value",
		},
		{
			name:    "unknown field gives a list of supported fields",
			in:      "spec.restartPolicy=Always",
			wantErr: "not supported",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := translatePodFieldSelector(tc.in)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result=%v)", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("translated params mismatch\nwant: %v\ngot:  %v", tc.want, got)
			}
		})
	}
}
