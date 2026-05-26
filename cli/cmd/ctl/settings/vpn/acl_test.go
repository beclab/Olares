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
		name     string
		tcp      []string
		udp      []string
		anyProto []string
		want     []AclInfo
	}{
		{name: "empty inputs produce empty payload", tcp: nil, udp: nil, anyProto: nil, want: []AclInfo{}},
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
		{
			name:     "any-proto only — empty proto entry preserved (Web 'Add ACL' parity)",
			anyProto: []string{"*:8080"},
			want:     []AclInfo{{Proto: "", Dst: []string{"*:8080"}}},
		},
		{
			name:     "any-proto sorted after tcp and udp",
			tcp:      []string{"*:80"},
			udp:      []string{"*:53"},
			anyProto: []string{"*:22"},
			want: []AclInfo{
				{Proto: "tcp", Dst: []string{"*:80"}},
				{Proto: "udp", Dst: []string{"*:53"}},
				{Proto: "", Dst: []string{"*:22"}},
			},
		},
		{
			name:     "any-proto trims and dedupes like tcp/udp",
			anyProto: []string{"  *:80 ", "*:80", "", "*:443"},
			want:     []AclInfo{{Proto: "", Dst: []string{"*:80", "*:443"}}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildACLPayload(c.tcp, c.udp, c.anyProto)
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("buildACLPayload(tcp=%v, udp=%v, anyProto=%v) = %#v, want %#v",
					c.tcp, c.udp, c.anyProto, got, c.want)
			}
		})
	}
}

func TestValidateACLDst(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "wildcard host port", in: "*:8080", wantErr: false},
		{name: "cidr host port", in: "192.168.1.0/24:22", wantErr: false},
		{name: "tagged host port", in: "tag:api:443", wantErr: false},
		{name: "host wildcard port", in: "example-host:*", wantErr: false},
		{name: "trims surrounding space", in: "  *:8080 ", wantErr: false},
		{name: "bare port rejected", in: "8080", wantErr: true},
		{name: "missing port rejected", in: "*:", wantErr: true},
		{name: "missing host rejected", in: ":8080", wantErr: true},
		{name: "empty string rejected", in: "", wantErr: true},
		{name: "whitespace-only rejected", in: "   ", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateACLDst(c.in)
			if (err != nil) != c.wantErr {
				t.Errorf("validateACLDst(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			}
		})
	}
}

func TestValidateACLPayload(t *testing.T) {
	if err := validateACLPayload(nil); err != nil {
		t.Errorf("nil payload should be valid, got %v", err)
	}
	good := []AclInfo{
		{Proto: "tcp", Dst: []string{"*:80", "*:443"}},
		{Proto: "udp", Dst: []string{"*:53"}},
		{Proto: "", Dst: []string{"*:22"}},
	}
	if err := validateACLPayload(good); err != nil {
		t.Errorf("good payload should be valid, got %v", err)
	}
	bad := []AclInfo{
		{Proto: "tcp", Dst: []string{"*:80", "8080"}}, // bare port slips through
	}
	err := validateACLPayload(bad)
	if err == nil {
		t.Fatal("expected error for bare-port dst, got nil")
	}
	if !contains(err.Error(), "8080") || !contains(err.Error(), "*:8080") {
		t.Errorf("error should cite offending value and a suggestion; got %q", err.Error())
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

// TestGetAllACLViaDoer covers all three wire shapes getAllACLViaDoer
// recognizes plus the two "empty" envelope responses. The flat BFL
// shape is the only one currently emitted on the wire (see
// framework/bfl/.../handle_headscale.go:handleHeadscaleACLList); the
// map and {name, acls} cases are kept as forward/back compat fallbacks
// and the tests here pin both so a future refactor can't silently
// regress them.
func TestGetAllACLViaDoer_FlatBFLShape(t *testing.T) {
	body := `{"code":0,"data":[
		{"appName":"halo","appOwner":"u1","proto":"tcp","dst":["x.x.x.x:123"]},
		{"appName":"halo","appOwner":"u1","proto":"udp","dst":["53"]},
		{"appName":"files","appOwner":"u1","proto":"tcp","dst":["80","443"]}
	]}`
	d := &envelopeFakeDoer{respondWith: []byte(body)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if d.method != "GET" || d.path != "/api/acl/all" {
		t.Errorf("call = %s %s, want GET /api/acl/all", d.method, d.path)
	}
	want := map[string][]AclInfo{
		"halo": {
			{Proto: "tcp", Dst: []string{"x.x.x.x:123"}},
			{Proto: "udp", Dst: []string{"53"}},
		},
		"files": {
			{Proto: "tcp", Dst: []string{"80", "443"}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %#v, want %#v", got, want)
	}
}

func TestGetAllACLViaDoer_SameProtoMultiRowUnion(t *testing.T) {
	// BFL today emits one row per (app, proto), but a legacy
	// spec.TailScaleACLs vector can splice in a second row. The CLI
	// merges them rather than stacking duplicate proto rows so the
	// table view stays one-row-per-proto.
	body := `{"code":0,"data":[
		{"appName":"halo","proto":"tcp","dst":["80","443"]},
		{"appName":"halo","proto":"tcp","dst":["443","8080"]}
	]}`
	d := &envelopeFakeDoer{respondWith: []byte(body)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]AclInfo{
		"halo": {{Proto: "tcp", Dst: []string{"80", "443", "8080"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %#v, want %#v", got, want)
	}
}

func TestGetAllACLViaDoer_MapShapeFallback(t *testing.T) {
	// Forward-compat: if BFL ever pivots to a map keyed by appName we
	// still decode it.
	body := `{"code":0,"data":{"halo":[{"proto":"tcp","dst":["80"]}]}}`
	d := &envelopeFakeDoer{respondWith: []byte(body)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]AclInfo{
		"halo": {{Proto: "tcp", Dst: []string{"80"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %#v, want %#v", got, want)
	}
}

func TestGetAllACLViaDoer_NameAclsFallback(t *testing.T) {
	// Historical shape: []{name, acls}. The flat-shape decode would
	// succeed structurally but every row has an empty appName, so we
	// fall through to this branch.
	body := `{"code":0,"data":[{"name":"halo","acls":[{"proto":"tcp","dst":["80"]}]}]}`
	d := &envelopeFakeDoer{respondWith: []byte(body)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]AclInfo{
		"halo": {{Proto: "tcp", Dst: []string{"80"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %#v, want %#v", got, want)
	}
}

func TestGetAllACLViaDoer_NonZeroCodeMeansEmpty(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":-1,"message":"not found"}`)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got = %#v, want empty map", got)
	}
}

func TestGetAllACLViaDoer_NullData(t *testing.T) {
	d := &envelopeFakeDoer{respondWith: []byte(`{"code":0,"data":null}`)}
	got, err := getAllACLViaDoer(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got = %#v, want empty map", got)
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
	// Empty/any-proto entries (what --any-proto and the Web UI's
	// "Add ACL" dialog emit) should render the proto column as
	// "any", not the empty string, so the success message stays
	// human-readable.
	got = summarizeACL([]AclInfo{
		{Proto: "", Dst: []string{"*:65001", "*:20"}},
	})
	if got != "any=*:20,*:65001" {
		t.Errorf("summarizeACL empty proto = %q, want any=*:20,*:65001", got)
	}
	got = summarizeACL([]AclInfo{
		{Proto: "tcp", Dst: []string{"*:80"}},
		{Proto: "", Dst: []string{"*:20"}},
	})
	if got != "tcp=*:80 any=*:20" {
		t.Errorf("summarizeACL tcp+empty = %q, want tcp=*:80 any=*:20", got)
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
