package vpn

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

// envelopeFakeDoer extends fakeDoer with a canned response body so we
// can exercise the get-then-merge-then-post paths. The recorded request
// is captured exactly like fakeDoer; the response is unmarshaled into
// `out` if non-nil so the bflEnvelope helpers can read it back.
type envelopeFakeDoer struct {
	fakeDoer
	respondWith []byte
}

func (e *envelopeFakeDoer) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
	e.method = method
	e.path = path
	e.body = body
	if e.wantErr != nil {
		return e.wantErr
	}
	if out != nil && len(e.respondWith) > 0 {
		return json.Unmarshal(e.respondWith, out)
	}
	return nil
}

func TestBuildACLPayload(t *testing.T) {
	cases := []struct {
		name string
		tcp  []string
		udp  []string
		want []AclInfo
	}{
		{name: "empty inputs produce empty payload", tcp: nil, udp: nil, want: []AclInfo{}},
		{
			name: "tcp only",
			tcp:  []string{"80", "443"},
			udp:  nil,
			want: []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}},
		},
		{
			name: "udp only",
			tcp:  nil,
			udp:  []string{"53"},
			want: []AclInfo{{Proto: "udp", Dst: []string{"53"}}},
		},
		{
			name: "both protocols, tcp before udp",
			tcp:  []string{"80"},
			udp:  []string{"53"},
			want: []AclInfo{
				{Proto: "tcp", Dst: []string{"80"}},
				{Proto: "udp", Dst: []string{"53"}},
			},
		},
		{
			name: "trims and dedupes",
			tcp:  []string{"  80 ", "80", "", "443"},
			udp:  []string{},
			want: []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}},
		},
		{
			name: "all-empty drops the proto entirely",
			tcp:  []string{"  ", ""},
			udp:  []string{"53"},
			want: []AclInfo{{Proto: "udp", Dst: []string{"53"}}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildACLPayload(c.tcp, c.udp)
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("buildACLPayload(tcp=%v, udp=%v) = %#v, want %#v", c.tcp, c.udp, got, c.want)
			}
		})
	}
}

func TestMergeACL(t *testing.T) {
	cases := []struct {
		name      string
		current   []AclInfo
		additions []AclInfo
		want      []AclInfo
	}{
		{
			name:      "empty current + empty additions → empty",
			current:   nil,
			additions: nil,
			want:      []AclInfo{},
		},
		{
			name: "additions only — appended in order",
			additions: []AclInfo{
				{Proto: "tcp", Dst: []string{"80"}},
				{Proto: "udp", Dst: []string{"53"}},
			},
			want: []AclInfo{
				{Proto: "tcp", Dst: []string{"80"}},
				{Proto: "udp", Dst: []string{"53"}},
			},
		},
		{
			name:    "current only — preserved as-is",
			current: []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			want:    []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
		},
		{
			name:      "shared proto unions dst preserving order",
			current:   []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}},
			additions: []AclInfo{{Proto: "tcp", Dst: []string{"80", "8080"}}},
			want:      []AclInfo{{Proto: "tcp", Dst: []string{"80", "443", "8080"}}},
		},
		{
			name:      "new proto appended at end of order",
			current:   []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			additions: []AclInfo{{Proto: "udp", Dst: []string{"53"}}},
			want: []AclInfo{
				{Proto: "tcp", Dst: []string{"80"}},
				{Proto: "udp", Dst: []string{"53"}},
			},
		},
		{
			name:      "proto comparison is case-insensitive",
			current:   []AclInfo{{Proto: "TCP", Dst: []string{"80"}}},
			additions: []AclInfo{{Proto: "tcp", Dst: []string{"443"}}},
			want:      []AclInfo{{Proto: "TCP", Dst: []string{"80", "443"}}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := mergeACL(c.current, c.additions)
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("mergeACL = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestSubtractACL(t *testing.T) {
	cases := []struct {
		name     string
		current  []AclInfo
		removals []AclInfo
		want     []AclInfo
	}{
		{name: "empty current ⇒ empty", current: nil, removals: nil, want: []AclInfo{}},
		{
			name:     "no removals ⇒ unchanged",
			current:  []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			removals: nil,
			want:     []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
		},
		{
			name:     "drop one dst keeps the proto",
			current:  []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}},
			removals: []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			want:     []AclInfo{{Proto: "tcp", Dst: []string{"443"}}},
		},
		{
			name:     "drop last dst removes the proto entirely",
			current:  []AclInfo{{Proto: "tcp", Dst: []string{"80"}}, {Proto: "udp", Dst: []string{"53"}}},
			removals: []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			want:     []AclInfo{{Proto: "udp", Dst: []string{"53"}}},
		},
		{
			name:     "removal of unknown proto is a no-op",
			current:  []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
			removals: []AclInfo{{Proto: "udp", Dst: []string{"53"}}},
			want:     []AclInfo{{Proto: "tcp", Dst: []string{"80"}}},
		},
		{
			name:     "proto comparison is case-insensitive",
			current:  []AclInfo{{Proto: "TCP", Dst: []string{"80", "443"}}},
			removals: []AclInfo{{Proto: "tcp", Dst: []string{"443"}}},
			want:     []AclInfo{{Proto: "TCP", Dst: []string{"80"}}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := subtractACL(c.current, c.removals)
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("subtractACL = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestPruneEmptyACL(t *testing.T) {
	in := []AclInfo{
		{Proto: "tcp", Dst: []string{"80", "  ", ""}},
		{Proto: "udp", Dst: []string{}},
		{Proto: "icmp", Dst: []string{"*"}},
	}
	got := pruneEmptyACL(in)
	want := []AclInfo{
		{Proto: "tcp", Dst: []string{"80"}},
		{Proto: "icmp", Dst: []string{"*"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pruneEmptyACL = %#v, want %#v", got, want)
	}
}

func TestPostAppACLViaDoer_BodyShape(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":0}`)}
	acls := []AclInfo{
		{Proto: "tcp", Dst: []string{"80", "443"}},
		{Proto: "udp", Dst: []string{}}, // pruned out
	}
	if err := postAppACLViaDoer(context.Background(), d, "my-app", acls); err != nil {
		t.Fatal(err)
	}
	if d.method != "POST" {
		t.Errorf("method = %q, want POST", d.method)
	}
	if d.path != "/api/acl/app/status" {
		t.Errorf("path = %q, want /api/acl/app/status", d.path)
	}

	raw, err := json.Marshal(d.body)
	if err != nil {
		t.Fatalf("re-marshal body: %v", err)
	}
	var got struct {
		Name string    `json:"name"`
		Acls []AclInfo `json:"acls"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Name != "my-app" {
		t.Errorf("body.name = %q, want my-app", got.Name)
	}
	want := []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}}
	if !reflect.DeepEqual(got.Acls, want) {
		t.Errorf("body.acls = %#v, want %#v", got.Acls, want)
	}
}

func TestPostAppACLViaDoer_RejectsUpstreamFailure(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":-1,"message":"app not found"}`)}
	err := postAppACLViaDoer(context.Background(), d, "ghost", []AclInfo{{Proto: "tcp", Dst: []string{"80"}}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	const want = "app not found"
	if !contains(err.Error(), want) {
		t.Errorf("err = %q, want containing %q", err.Error(), want)
	}
}

func TestGetAppACLViaDoer_DecodesEnvelope(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":0,"data":[{"proto":"tcp","dst":["80","443"]}]}`)}
	got, err := getAppACLViaDoer(context.Background(), d, "my-app")
	if err != nil {
		t.Fatal(err)
	}
	if d.path != "/api/acl/app/status?name=my-app" {
		t.Errorf("path = %q", d.path)
	}
	want := []AclInfo{{Proto: "tcp", Dst: []string{"80", "443"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %#v, want %#v", got, want)
	}
}

func TestGetAppACLViaDoer_NonZeroCodeMeansEmpty(t *testing.T) {
	// Mirrors the SPA's behavior: a non-zero `code` for this URL is
	// "no ACL configured" rather than a hard error.
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":-1,"message":"not found"}`)}
	got, err := getAppACLViaDoer(context.Background(), d, "ghost")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got = %#v, want empty", got)
	}
}

func TestGetAppACLViaDoer_QueryEscaping(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":0,"data":[]}`)}
	if _, err := getAppACLViaDoer(context.Background(), d, "weird name & co"); err != nil {
		t.Fatal(err)
	}
	if d.path != "/api/acl/app/status?name=weird+name+%26+co" {
		t.Errorf("path = %q", d.path)
	}
}

func TestSummarizeACL(t *testing.T) {
	got := summarizeACL([]AclInfo{
		{Proto: "TCP", Dst: []string{"443", "80"}},
		{Proto: "udp", Dst: []string{"53"}},
	})
	const want = "tcp=443,80 udp=53"
	if got != want {
		t.Errorf("summarizeACL = %q, want %q", got, want)
	}
	if got := summarizeACL(nil); got != "(empty)" {
		t.Errorf("summarizeACL(nil) = %q, want (empty)", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (sub == "" || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
