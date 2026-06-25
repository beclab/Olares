package access

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
)

func TestClassifyNetErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want NetErrKind
	}{
		{"nil", nil, KindNone},
		{"caller-cancel", context.Canceled, KindCallerCancel},
		{"deadline", context.DeadlineExceeded, KindTimeout},
		{"dns-notfound", &net.DNSError{Err: "no such host", IsNotFound: true}, KindNameNotFound},
		{"dns-timeout", &net.DNSError{Err: "timeout", IsTimeout: true}, KindTimeout},
		{"dns-servfail", &net.DNSError{Err: "server misbehaving"}, KindDNSUnreachable},
		{"refused", syscall.ECONNREFUSED, KindRefused},
		{"net-unreach", syscall.ENETUNREACH, KindLocalNetDown},
		{"host-unreach", syscall.EHOSTUNREACH, KindLocalNetDown},
		{"tls-record", tls.RecordHeaderError{Msg: "bad"}, KindTLS},
		{"tls-string", errors.New("tls: handshake failure"), KindTLS},
		{"x509-string", errors.New("x509: certificate signed by unknown authority"), KindTLS},
		{"other", errors.New("something weird"), KindOther},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyNetErr(c.err); got != c.want {
				t.Errorf("classifyNetErr(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

func TestClassifyNetErrWrapped(t *testing.T) {
	wrapped := fmt.Errorf("dial tcp: %w", syscall.ECONNREFUSED)
	if got := classifyNetErr(wrapped); got != KindRefused {
		t.Errorf("wrapped ECONNREFUSED classified as %d, want %d", got, KindRefused)
	}
}

func TestIsSwitchableNetErr(t *testing.T) {
	switchable := []error{
		&net.DNSError{IsNotFound: true},
		&net.DNSError{Err: "servfail"},
		syscall.ENETUNREACH,
		syscall.ECONNREFUSED,
		context.DeadlineExceeded,
	}
	for _, err := range switchable {
		if !IsSwitchableNetErr(err, context.Background()) {
			t.Errorf("expected %v to be switchable", err)
		}
	}

	notSwitchable := []error{
		nil,
		errors.New("tls: handshake failure"),
		context.Canceled,
		errors.New("some opaque error"),
	}
	for _, err := range notSwitchable {
		if IsSwitchableNetErr(err, context.Background()) {
			t.Errorf("expected %v NOT to be switchable", err)
		}
	}
}

func TestIsSwitchableNetErrCanceledContextShortCircuits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Even a normally-switchable error must not switch once the caller gave up.
	if IsSwitchableNetErr(syscall.ECONNREFUSED, ctx) {
		t.Error("a canceled context should make any error non-switchable")
	}
}

func TestUnreachableErrorMessage(t *testing.T) {
	local := &UnreachableError{OlaresID: "alice@olares.com", LastKind: KindLocalNetDown}
	if !strings.Contains(local.Error(), "local network") {
		t.Errorf("local-down message = %q, want a local-network hint", local.Error())
	}
	remote := &UnreachableError{OlaresID: "alice@olares.com", LastKind: KindRefused}
	if !strings.Contains(remote.Error(), "unreachable from here") {
		t.Errorf("remote message = %q", remote.Error())
	}
	// Empty OlaresID still produces a sensible subject.
	anon := &UnreachableError{LastKind: KindRefused}
	if !strings.Contains(anon.Error(), "the Olares instance") {
		t.Errorf("anon message = %q", anon.Error())
	}
}

func TestIsUnreachable(t *testing.T) {
	ue := &UnreachableError{LastKind: KindOther}
	if !IsUnreachable(ue) {
		t.Error("IsUnreachable should match *UnreachableError")
	}
	if !IsUnreachable(fmt.Errorf("wrap: %w", ue)) {
		t.Error("IsUnreachable should match a wrapped *UnreachableError")
	}
	if IsUnreachable(errors.New("plain")) {
		t.Error("IsUnreachable should not match a plain error")
	}
}

func TestNewUnreachable(t *testing.T) {
	cause := fmt.Errorf("dial tcp 1.2.3.4:443: %w", syscall.ECONNREFUSED)
	ue := NewUnreachable("alice@olares.com", cause)

	if ue.OlaresID != "alice@olares.com" {
		t.Errorf("OlaresID = %q, want alice@olares.com", ue.OlaresID)
	}
	// LastKind is derived from classifying the cause.
	if ue.LastKind != KindRefused {
		t.Errorf("LastKind = %d, want KindRefused (%d)", ue.LastKind, KindRefused)
	}
	// Unwrap exposes the cause so typed checks chain through.
	if !errors.Is(ue, syscall.ECONNREFUSED) {
		t.Error("errors.Is(NewUnreachable(...), ECONNREFUSED) should be true")
	}
	if errors.Unwrap(ue) != cause {
		t.Errorf("Unwrap() = %v, want the original cause", errors.Unwrap(ue))
	}
	// Still satisfies IsUnreachable and renders the friendly message.
	if !IsUnreachable(ue) {
		t.Error("NewUnreachable result should satisfy IsUnreachable")
	}
	if !strings.Contains(ue.Error(), "unreachable from here") {
		t.Errorf("message = %q, want the friendly remote-unreachable text", ue.Error())
	}
}

func TestNewUnreachableLocalNetDownMessage(t *testing.T) {
	ue := NewUnreachable("alice@olares.com", syscall.ENETUNREACH)
	if ue.LastKind != KindLocalNetDown {
		t.Errorf("LastKind = %d, want KindLocalNetDown (%d)", ue.LastKind, KindLocalNetDown)
	}
	if !strings.Contains(ue.Error(), "local network") {
		t.Errorf("message = %q, want a local-network hint", ue.Error())
	}
}
