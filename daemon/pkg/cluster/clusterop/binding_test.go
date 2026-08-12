package clusterop

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func at(ms int64) time.Time { return time.UnixMilli(ms).UTC() }

// signedBody is what CheckJWS hands back: the "body" member of the payload the
// owner's key covered, decoded as generic JSON.
func signedBody(t *testing.T, raw string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("bad test fixture %s: %v", raw, err)
	}
	return v
}

func bindingCode(t *testing.T, err error) string {
	t.Helper()
	var be *BindingError
	if !errors.As(err, &be) {
		t.Fatalf("error %v is not a BindingError, so no caller can map it to a stable code", err)
	}
	return be.Code
}

// The binding is carried in the free-form "body" TermiPass already signs, so
// adding it changes no signing primitive.
func TestParseBindingReadsWhatTheOwnerSigned(t *testing.T) {
	body := signedBody(t, `{
		"username": "alice",
		"clusterId": "cluster-1",
		"type": "reboot",
		"requestId": "client-1",
		"scope": "cluster",
		"expiresAt": 1700000600000
	}`)

	b, err := ParseBinding(body)
	if err != nil {
		t.Fatalf("ParseBinding: %v", err)
	}
	want := Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster, ExpiresAt: 1700000600000}
	if b != want {
		t.Errorf("binding = %+v, want %+v", b, want)
	}
}

// A route that answers for an explicit module set reads the signature
// against that same set. Without it a request for a type only that set holds
// would be refused as unbound before the route ever saw it — and, worse, the
// two could disagree about which operation a signature authorized.
func TestParseBindingInReadsATypeOnlyTheGivenModuleSetHolds(t *testing.T) {
	registry := registryWith(t, registryTestModule{typ: Type("bake-cake")})
	body := signedBody(t, `{
		"clusterId": "cluster-1",
		"type": "bake-cake",
		"requestId": "client-1",
		"scope": "cluster",
		"expiresAt": 1700000600000
	}`)

	b, err := ParseBindingIn(registry, body)
	if err != nil {
		t.Fatalf("ParseBindingIn: %v", err)
	}
	if b.Type != Type("bake-cake") {
		t.Errorf("type = %q, want the type the given module set holds", b.Type)
	}
	if _, err := ParseBinding(body); err == nil {
		t.Error("the daemon-wide module set accepted a type nothing registered into it")
	}
}

func TestParseBindingRefusesMissingClusterID(t *testing.T) {
	_, err := ParseBinding(signedBody(t, `{
		"type": "reboot",
		"requestId": "client-1",
		"scope": "cluster",
		"expiresAt": 1700000600000
	}`))
	if err == nil {
		t.Fatal("binding without cluster id was accepted")
	}
	if got := bindingCode(t, err); got != CodeSignatureUnbound {
		t.Errorf("code = %q, want %s", got, CodeSignatureUnbound)
	}
}

func TestNodeBindingIncludesTheTargetNode(t *testing.T) {
	body := signedBody(t, `{
		"clusterId": "cluster-1",
		"type": "reboot",
		"requestId": "client-1",
		"scope": "node",
		"target": "worker-1",
		"expiresAt": 1700000600000
	}`)

	b, err := ParseBinding(body)
	if err != nil {
		t.Fatalf("ParseBinding: %v", err)
	}
	want := Binding{
		ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeNode,
		Target: "worker-1", ExpiresAt: 1700000600000,
	}
	if b != want {
		t.Errorf("binding = %+v, want %+v", b, want)
	}
	if err := b.Authorizes(Binding{
		ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeNode, Target: "worker-2",
	}, at(1700000000000)); bindingCode(t, err) != CodeSignatureMismatch {
		t.Errorf("another node was authorized: %v", err)
	}
}

// A signature that says nothing about the operation is the whole problem: it
// is a twenty-minute bearer for every dangerous route at once.
func TestParseBindingRefusesASignatureThatBindsNothing(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"an ordinary single-node command body", `{"username":"alice"}`},
		{"no body at all", `null`},
		{"a body that is not an object", `"alice"`},
		{"no type", `{"requestId":"client-1","scope":"cluster","expiresAt":1700000600000}`},
		{"no request id", `{"type":"reboot","scope":"cluster","expiresAt":1700000600000}`},
		{"no scope", `{"type":"reboot","requestId":"client-1","expiresAt":1700000600000}`},
		{"no expiry", `{"type":"reboot","requestId":"client-1","scope":"cluster"}`},
		{"a type this daemon cannot perform", `{"type":"halt","requestId":"c","scope":"cluster","expiresAt":1700000600000}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseBinding(signedBody(t, tc.body))
			if err == nil {
				t.Fatal("an unbound signature was accepted")
			}
			if got := bindingCode(t, err); got != CodeSignatureUnbound {
				t.Errorf("code = %q, want %s", got, CodeSignatureUnbound)
			}
		})
	}
}

func TestBindingAuthorizesTheRequestItNames(t *testing.T) {
	b := Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster, ExpiresAt: 1700000600000}
	want := Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster}

	if err := b.Authorizes(want, at(1700000000000)); err != nil {
		t.Fatalf("a matching request was refused: %v", err)
	}
}

// This is the point of the binding: a signature the owner produced for one
// dangerous thing cannot be replayed at another.
func TestBindingRefusesADifferentRequest(t *testing.T) {
	b := Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster, ExpiresAt: 1700000600000}

	for _, tc := range []struct {
		name string
		want Binding
	}{
		{"another cluster", Binding{ClusterID: "cluster-2", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster}},
		{"another operation type", Binding{ClusterID: "cluster-1", Type: TypeShutdown, RequestID: "client-1", Scope: ScopeCluster}},
		{"another request id", Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-2", Scope: ScopeCluster}},
		{"another scope", Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: "node"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := b.Authorizes(tc.want, at(1700000000000))
			if err == nil {
				t.Fatal("a signature bound to something else was accepted")
			}
			if got := bindingCode(t, err); got != CodeSignatureMismatch {
				t.Errorf("code = %q, want %s", got, CodeSignatureMismatch)
			}
		})
	}
}

func TestBindingExpires(t *testing.T) {
	b := Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster, ExpiresAt: 1700000600000}

	err := b.Authorizes(Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster}, at(1700000600001))
	if err == nil {
		t.Fatal("an expired signature was accepted")
	}
	if got := bindingCode(t, err); got != CodeSignatureExpired {
		t.Errorf("code = %q, want %s", got, CodeSignatureExpired)
	}
}

// The expiry is the signer's, so it is also the signer's to abuse. A grant
// good for a year is a permanent key to power the cluster off.
func TestBindingRefusesAnExpiryFurtherOutThanTheDaemonAllows(t *testing.T) {
	now := at(1700000000000)
	b := Binding{
		ClusterID: "cluster-1",
		Type:      TypeReboot,
		RequestID: "client-1",
		Scope:     ScopeCluster,
		ExpiresAt: now.Add(MaxSignatureLifetime + time.Minute).UnixMilli(),
	}

	err := b.Authorizes(Binding{ClusterID: "cluster-1", Type: TypeReboot, RequestID: "client-1", Scope: ScopeCluster}, now)
	if err == nil {
		t.Fatal("a signature valid for longer than the daemon permits was accepted")
	}
	if got := bindingCode(t, err); got != CodeSignatureExpired {
		t.Errorf("code = %q, want %s", got, CodeSignatureExpired)
	}
}

// A BindingError is what reaches an HTTP client, so it must carry nothing but
// the stable code and a fixed sentence.
func TestBindingErrorsSayNothingAboutTheSignature(t *testing.T) {
	_, err := ParseBinding(signedBody(t, `{"username":"alice","secret":"hunter2"}`))
	var be *BindingError
	if !errors.As(err, &be) {
		t.Fatalf("error %v is not a BindingError", err)
	}
	if be.Message == "" {
		t.Error("a refusal with no message tells the caller nothing")
	}
	if got := be.Error(); got == "" || got == be.Code {
		t.Errorf("Error() = %q", got)
	}
}
