package clusterop

import (
	"encoding/json"
	"fmt"
	"time"
)

// Scope names what a signature is allowed to act on. It exists so that a
// signature produced to power one node cannot be presented to power the whole
// cluster, and vice versa.
const (
	ScopeCluster = "cluster"
	ScopeNode    = "node"
)

// MaxSignatureLifetime bounds how long the signer may declare its own
// signature valid for. The expiry is chosen by whoever signs, so without a
// ceiling a client could mint a grant good for a year, which is a permanent
// key to power the cluster off.
const MaxSignatureLifetime = 10 * time.Minute

// Stable codes for a signature that does not authorize the request carrying
// it. They are distinct on purpose: unbound means the signature never said
// what it was for, mismatch means it said something else.
const (
	CodeSignatureUnbound     = "signature_unbound"
	CodeSignatureMismatch    = "signature_mismatch"
	CodeSignatureExpired     = "signature_expired"
	CodeUnsupportedOperation = "unsupported_operation"
	CodeInvalidRequest       = "invalid_request"
)

// BindingError refuses a request on the evidence it presented. Message is
// fixed text: it reaches an HTTP client, and a refusal that described the
// signature would tell an attacker which part of a forgery to correct.
type BindingError struct {
	Code    string
	Message string
}

func (e *BindingError) Error() string { return e.Code + ": " + e.Message }

// Binding is what the owner signed about the operation, carried in the
// free-form body TermiPass already puts in the JWS payload. Adding it changes
// no signing primitive: the wire is
//
//	{"did":…,"name":…,"time":…,"body":{
//	   "type":"reboot","requestId":"client-1",
//	   "scope":"cluster","expiresAt":1700000600000}}
//
// and the daemon verifies the signature exactly as it does today, then checks
// that these fields describe the request that arrived. Node scope also carries
// the target node name.
type Binding struct {
	ClusterID string
	Type      Type
	RequestID string
	Scope     string
	Target    string

	// ExpiresAt is Unix milliseconds, matching the "time" field already in the
	// payload rather than introducing a second time format.
	ExpiresAt int64
}

// bindingWire is the signed body as it is written. It is decoded rather than
// asserted field by field so that a wrong type reads as an absent field.
type bindingWire struct {
	ClusterID string `json:"clusterId"`
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Scope     string `json:"scope"`
	Target    string `json:"target"`
	ExpiresAt int64  `json:"expiresAt"`

	// Body is the nested member CheckJWS hands back when it returns the whole
	// payload rather than only what was under "body".
	Body *bindingWire `json:"body"`
}

// ParseBinding reads the binding out of a verified signature's signed body.
//
// A body that says nothing about an operation is refused rather than treated
// as a general grant: an unbound owner signature is a bearer token for every
// dangerous route at once, for as long as the daemon's clock skew allows.
func ParseBinding(signed any) (Binding, error) {
	unbound := &BindingError{
		Code:    CodeSignatureUnbound,
		Message: "the signature does not authorize this operation",
	}

	raw, err := json.Marshal(signed)
	if err != nil {
		return Binding{}, unbound
	}
	var wire bindingWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return Binding{}, unbound
	}
	if wire.Type == "" && wire.Body != nil {
		wire = *wire.Body
	}
	if wire.ClusterID == "" || wire.RequestID == "" || wire.Scope == "" || wire.ExpiresAt == 0 ||
		(wire.Scope == ScopeNode && wire.Target == "") {
		return Binding{}, unbound
	}
	opType, err := ParseType(wire.Type)
	if err != nil {
		return Binding{}, unbound
	}
	return Binding{
		ClusterID: wire.ClusterID,
		Type:      opType,
		RequestID: wire.RequestID,
		Scope:     wire.Scope,
		Target:    wire.Target,
		ExpiresAt: wire.ExpiresAt,
	}, nil
}

// Authorizes reports whether this signature covers the request that arrived.
// want carries no expiry: the request describes itself, the signature says
// until when.
func (b Binding) Authorizes(want Binding, now time.Time) error {
	if (want.ClusterID != "" && b.ClusterID != want.ClusterID) ||
		b.Type != want.Type || b.RequestID != want.RequestID ||
		b.Scope != want.Scope || b.Target != want.Target {
		return &BindingError{
			Code:    CodeSignatureMismatch,
			Message: "the signature authorizes a different operation",
		}
	}

	expiry := time.UnixMilli(b.ExpiresAt)
	if !now.Before(expiry) {
		return &BindingError{Code: CodeSignatureExpired, Message: "the signature has expired"}
	}
	if expiry.After(now.Add(MaxSignatureLifetime)) {
		return &BindingError{
			Code: CodeSignatureExpired,
			Message: fmt.Sprintf("a signature may not be valid for longer than %s",
				MaxSignatureLifetime),
		}
	}
	return nil
}
