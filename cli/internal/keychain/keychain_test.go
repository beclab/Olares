package keychain

import (
	"errors"
	"strings"
	"testing"
)

// TestWrapError_PassThroughs locks down the two cases where wrapError must
// NOT wrap: nil errors and ErrNotFound. A drift here would either spam
// callers with bogus failures or break errors.Is checks.
func TestWrapError_PassThroughs(t *testing.T) {
	if got := wrapError("Get", "svc", "acct", nil); got != nil {
		t.Errorf("wrapError(nil) = %v; want nil", got)
	}
	if got := wrapError("Get", "svc", "acct", ErrNotFound); !errors.Is(got, ErrNotFound) {
		t.Errorf("wrapError(ErrNotFound) lost the sentinel: %v", got)
	}
}

// TestWrapError_TerseDefault checks the default-mode invariant: the message
// is short, names the (service, account) pair, includes the cause, and does
// NOT include the long English hint the old format always emitted.
func TestWrapError_TerseDefault(t *testing.T) {
	prev := debugLookup
	debugLookup = func() bool { return false }
	defer func() { debugLookup = prev }()

	cause := errors.New("boom")
	got := wrapError("Get", "olares-cli", "alice@olares.com", cause)
	msg := got.Error()

	if !strings.Contains(msg, "olares-cli/alice@olares.com") {
		t.Errorf("missing service/account in terse message: %q", msg)
	}
	if !strings.Contains(msg, "boom") {
		t.Errorf("missing cause in terse message: %q", msg)
	}
	if strings.Contains(msg, "OS keychain / credential manager is locked") {
		t.Errorf("verbose hint leaked into terse message: %q", msg)
	}
	if !errors.Is(got, cause) {
		t.Errorf("wrapped error lost cause via errors.Is: %v", got)
	}
}

// TestWrapError_DebugVerbose flips the seam to the debug branch and
// confirms the long hint appears AND that errNotInitialized triggers the
// dedicated re-login hint instead of the generic one.
func TestWrapError_DebugVerbose(t *testing.T) {
	prev := debugLookup
	debugLookup = func() bool { return true }
	defer func() { debugLookup = prev }()

	cause := errors.New("boom")
	msg := wrapError("Set", "olares-cli", "alice@olares.com", cause).Error()
	if !strings.Contains(msg, "OS keychain / credential manager is locked") {
		t.Errorf("expected generic hint in debug mode, got: %q", msg)
	}

	msg2 := wrapError("Get", "olares-cli", "alice@olares.com", errNotInitialized).Error()
	if !strings.Contains(msg2, "master key may have been deleted or corrupted") {
		t.Errorf("expected errNotInitialized hint, got: %q", msg2)
	}
	if strings.Contains(msg2, "OS keychain / credential manager is locked") {
		t.Errorf("generic hint leaked into errNotInitialized branch: %q", msg2)
	}
}
