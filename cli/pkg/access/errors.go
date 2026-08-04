// Package access owns olares-cli's "where am I relative to this Olares?"
// machinery: building http.Transports for a given network position
// (Location), probing which position is reachable, and classifying the
// transport-layer errors that drive an automatic re-probe.
//
// It depends only on pkg/olares (for Location + URL derivation), so it can be
// imported by auth, credential, cmdutil and whoami without import cycles.
package access

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
)

// NetErrKind is a coarse classification of a transport-layer error. It exists
// so the reprobe path can decide whether a failure is worth switching
// Location for, and so user-facing messages can lean toward "your local
// network is down" vs "Olares is unreachable" — all heuristically, WITHOUT
// any third-party network probing.
type NetErrKind int

const (
	// KindNone is the zero value: err == nil.
	KindNone NetErrKind = iota
	// KindNameNotFound is an authoritative "no such host" (DNS NXDOMAIN).
	KindNameNotFound
	// KindDNSUnreachable is a DNS lookup that failed for a non-NXDOMAIN
	// reason (resolver unreachable, SERVFAIL, temporary failure).
	KindDNSUnreachable
	// KindLocalNetDown is ENETUNREACH / EHOSTUNREACH — the local network
	// stack or route is down, before we ever reached a peer.
	KindLocalNetDown
	// KindRefused is ECONNREFUSED — we reached the host but the port is
	// closed (a peer answered).
	KindRefused
	// KindTimeout is an i/o timeout / deadline exceeded on the dial or read.
	KindTimeout
	// KindTLS is a TLS handshake / certificate verification failure (we
	// reached a TLS endpoint, so the network itself is fine).
	KindTLS
	// KindCallerCancel is the caller's context being canceled — not a
	// network condition, never switchable.
	KindCallerCancel
	// KindOther is anything we couldn't confidently bucket.
	KindOther
)

// classifyNetErr buckets a transport error into a NetErrKind. The order of
// checks matters: caller-cancel and deadline are checked before the typed
// network errors so a ctx-driven failure isn't mistaken for a peer condition.
func classifyNetErr(err error) NetErrKind {
	if err == nil {
		return KindNone
	}
	if errors.Is(err, context.Canceled) {
		return KindCallerCancel
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return KindTimeout
	}

	// TLS: we reached a TLS endpoint, so the path is up; never switchable.
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return KindTLS
	}
	var recErr tls.RecordHeaderError
	if errors.As(err, &recErr) {
		return KindTLS
	}

	// DNS.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		switch {
		case dnsErr.IsNotFound:
			return KindNameNotFound
		case dnsErr.IsTimeout:
			return KindTimeout
		default:
			return KindDNSUnreachable
		}
	}

	// Raw syscall errnos.
	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return KindRefused
	case errors.Is(err, syscall.ENETUNREACH), errors.Is(err, syscall.EHOSTUNREACH):
		return KindLocalNetDown
	}

	// Generic net.Error timeout (covers some platform dial timeouts that
	// don't surface as a typed DNSError or deadline).
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return KindTimeout
	}

	// Last-ditch string sniff for TLS errors that don't expose a typed form
	// on every Go/platform combination.
	msg := err.Error()
	if strings.Contains(msg, "tls:") || strings.Contains(msg, "x509") || strings.Contains(msg, "certificate") {
		return KindTLS
	}
	return KindOther
}

// IsSwitchableNetErr reports whether err is a transport-layer failure worth
// re-probing the Location for. A canceled caller context short-circuits to
// false (the user gave up; don't paper over it with a slow re-probe). TLS
// failures and anything with an HTTP response are NOT switchable — the
// network path is clearly up, the problem is elsewhere.
func IsSwitchableNetErr(err error, ctx context.Context) bool {
	if ctx != nil && ctx.Err() != nil {
		return false
	}
	switch classifyNetErr(err) {
	case KindNameNotFound, KindDNSUnreachable, KindLocalNetDown, KindRefused, KindTimeout:
		return true
	default:
		return false
	}
}

// UnreachableError is returned by ProbeLocation when every connection method
// failed. LastKind carries the classification of the final (external) probe
// failure so callers can tailor the message; it is purely a hint. Cause, when
// set, is the underlying transport error and is exposed via Unwrap so typed
// checks (errors.Is(err, syscall.ECONNREFUSED), etc.) still chain through.
type UnreachableError struct {
	OlaresID string
	LastKind NetErrKind
	Cause    error
}

func (e *UnreachableError) Error() string {
	id := e.OlaresID
	if id == "" {
		id = "the Olares instance"
	}
	switch e.LastKind {
	case KindLocalNetDown, KindDNSUnreachable, KindTimeout:
		return fmt.Sprintf("could not reach %s — your local network may be down or offline", id)
	default:
		return fmt.Sprintf("%s is unreachable from here (tried LAN, intranet, and public access)", id)
	}
}

// Unwrap exposes the underlying transport error (when any) so errors.Is / As
// keep working against the original cause.
func (e *UnreachableError) Unwrap() error { return e.Cause }

// NewUnreachable builds an *UnreachableError from a concrete transport failure,
// classifying cause into LastKind and retaining it as the wrapped Cause. Use
// this when the friendly "unreachable" message should be surfaced to the user
// but the raw error must stay reachable (e.g. the runtime reprobe path that
// couldn't switch to a working connection method).
func NewUnreachable(olaresID string, cause error) *UnreachableError {
	return &UnreachableError{
		OlaresID: olaresID,
		LastKind: classifyNetErr(cause),
		Cause:    cause,
	}
}

// IsUnreachable reports whether err is, or wraps, an *UnreachableError.
func IsUnreachable(err error) bool {
	var ue *UnreachableError
	return errors.As(err, &ue)
}
