package vpn

import (
	"context"
	"reflect"
	"testing"
)

// fakeDoer captures the last DoJSON call so tests can assert exactly
// what would have hit the wire. It returns nilErr by default; tests
// can flip wantErr to simulate transport failures.
type fakeDoer struct {
	method  string
	path    string
	body    interface{}
	wantErr error
}

func (f *fakeDoer) DoJSON(_ context.Context, method, path string, body, _ interface{}) error {
	f.method = method
	f.path = path
	f.body = body
	return f.wantErr
}

func TestNormalizeHeadscaleTags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty", in: []string{}, want: nil},
		{
			name: "bare names get tag: prefix",
			in:   []string{"ops", "laptop"},
			want: []string{"tag:ops", "tag:laptop"},
		},
		{
			name: "already-prefixed pass through",
			in:   []string{"tag:ops", "tag:laptop"},
			want: []string{"tag:ops", "tag:laptop"},
		},
		{
			name: "mixed input is normalized",
			in:   []string{"ops", "tag:laptop"},
			want: []string{"tag:ops", "tag:laptop"},
		},
		{
			name: "trims and drops empties",
			in:   []string{"  ops  ", "", "   ", "laptop"},
			want: []string{"tag:ops", "tag:laptop"},
		},
		{
			name: "dedupes preserving first-seen order",
			in:   []string{"ops", "laptop", "tag:ops", "ops", "laptop"},
			want: []string{"tag:ops", "tag:laptop"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := normalizeHeadscaleTags(c.in)
			// Treat nil and zero-length slice as equal — the wire
			// representation is the same: an empty JSON array.
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("normalizeHeadscaleTags(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestDoRenameViaDoer_PathEscaping(t *testing.T) {
	d := &fakeDoer{}
	if err := doRenameViaDoer(context.Background(), d, "7", "alice's laptop"); err != nil {
		t.Fatal(err)
	}
	if d.method != "POST" {
		t.Errorf("method = %q, want POST", d.method)
	}
	want := "/headscale/machine/7/rename/alice%27s%20laptop"
	if d.path != want {
		t.Errorf("path = %q, want %q", d.path, want)
	}
	if d.body != nil {
		t.Errorf("body = %v, want nil", d.body)
	}
}

func TestDoDeleteViaDoer(t *testing.T) {
	d := &fakeDoer{}
	if err := doDeleteViaDoer(context.Background(), d, "7"); err != nil {
		t.Fatal(err)
	}
	if d.method != "DELETE" {
		t.Errorf("method = %q, want DELETE", d.method)
	}
	if d.path != "/headscale/machine/7" {
		t.Errorf("path = %q, want /headscale/machine/7", d.path)
	}
	if d.body != nil {
		t.Errorf("body = %v, want nil", d.body)
	}
}

func TestDoTagsSetViaDoer_WrapsTags(t *testing.T) {
	d := &fakeDoer{}
	wire := normalizeHeadscaleTags([]string{"ops", "tag:laptop"})
	if err := doTagsSetViaDoer(context.Background(), d, "7", wire); err != nil {
		t.Fatal(err)
	}
	if d.method != "POST" {
		t.Errorf("method = %q, want POST", d.method)
	}
	if d.path != "/headscale/machine/7/tags" {
		t.Errorf("path = %q, want /headscale/machine/7/tags", d.path)
	}
	body, ok := d.body.(map[string][]string)
	if !ok {
		t.Fatalf("body type = %T, want map[string][]string", d.body)
	}
	want := []string{"tag:ops", "tag:laptop"}
	if !reflect.DeepEqual(body["tags"], want) {
		t.Errorf("body.tags = %v, want %v", body["tags"], want)
	}
}

func TestDoTagsSetViaDoer_ClearsWithEmptySlice(t *testing.T) {
	d := &fakeDoer{}
	wire := normalizeHeadscaleTags(nil)
	if err := doTagsSetViaDoer(context.Background(), d, "7", wire); err != nil {
		t.Fatal(err)
	}
	body, ok := d.body.(map[string][]string)
	if !ok {
		t.Fatalf("body type = %T, want map[string][]string", d.body)
	}
	if len(body["tags"]) != 0 {
		t.Errorf("body.tags = %v, want empty", body["tags"])
	}
}

func TestDoRouteToggleViaDoer(t *testing.T) {
	cases := []struct {
		enable bool
		want   string
	}{
		{true, "/headscale/routes/12/enable"},
		{false, "/headscale/routes/12/disable"},
	}
	for _, c := range cases {
		d := &fakeDoer{}
		if err := doRouteToggleViaDoer(context.Background(), d, "12", c.enable); err != nil {
			t.Fatal(err)
		}
		if d.path != c.want {
			t.Errorf("enable=%v: path = %q, want %q", c.enable, d.path, c.want)
		}
		if d.method != "POST" {
			t.Errorf("enable=%v: method = %q, want POST", c.enable, d.method)
		}
		if _, ok := d.body.(struct{}); !ok {
			t.Errorf("enable=%v: body = %T, want struct{}", c.enable, d.body)
		}
	}
}

func TestResolvePolicyFlag(t *testing.T) {
	cases := []struct {
		denyAll, allowAll bool
		wantValue         int
		wantLabel         string
		wantErr           bool
	}{
		{false, false, 0, "", true},  // neither
		{true, true, 0, "", true},    // both
		{true, false, 1, "deny-all", false},
		{false, true, 0, "allow-all", false},
	}
	for _, c := range cases {
		val, label, err := resolvePolicyFlag(c.denyAll, c.allowAll)
		if (err != nil) != c.wantErr {
			t.Errorf("deny=%v allow=%v: err = %v, wantErr = %v", c.denyAll, c.allowAll, err, c.wantErr)
			continue
		}
		if c.wantErr {
			continue
		}
		if val != c.wantValue {
			t.Errorf("deny=%v allow=%v: value = %d, want %d", c.denyAll, c.allowAll, val, c.wantValue)
		}
		if label != c.wantLabel {
			t.Errorf("deny=%v allow=%v: label = %q, want %q", c.denyAll, c.allowAll, label, c.wantLabel)
		}
	}
}

func TestDoPolicySetViaDoer(t *testing.T) {
	d := &fakeDoer{}
	if err := doPolicySetViaDoer(context.Background(), d, 1); err != nil {
		t.Fatal(err)
	}
	if d.method != "POST" {
		t.Errorf("method = %q, want POST", d.method)
	}
	if d.path != "/api/launcher-public-domain-access-policy" {
		t.Errorf("path = %q", d.path)
	}
	body, ok := d.body.(publicDomainPolicy)
	if !ok {
		t.Fatalf("body type = %T, want publicDomainPolicy", d.body)
	}
	if body.DenyAll != 1 {
		t.Errorf("body.DenyAll = %d, want 1", body.DenyAll)
	}
}
