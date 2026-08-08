package market

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// Market answers a cancel it will not perform with one 404 and one sentence,
// whether the app does not exist or its operation has already settled. These
// pin which of the two explainCancelRejection is willing to name, since
// guessing wrong sends a reader after a misspelt name or an old backend when
// the real answer is "the install you were cancelling already finished".

func TestSettledStateCancelExplainsItselfAndKeepsTheOriginalText(t *testing.T) {
	wire := &APIError{
		StatusCode: http.StatusNotFound,
		Message:    "App not found or current state does not allow operation",
	}
	got := explainCancelRejection(wire, "jellyfin", "market-local", &installedAppRow{State: "running", Source: "market-local"})

	if errors.Is(got, wire) == false {
		t.Fatalf("the wire error must stay in the chain, got %v", got)
	}
	text := got.Error()
	for _, want := range []string{
		"API error (HTTP 404)",
		"was in state 'running'",
		"olares-cli market status jellyfin",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}

func TestCancelRejectionIsLeftAloneWhenTheAppWasNeverFound(t *testing.T) {
	// No state row means "not found" is the half that plausibly applied, and
	// inventing a settled state the CLI never saw would be worse than silence.
	wire := &APIError{StatusCode: http.StatusNotFound, Message: "App not found"}
	if got := explainCancelRejection(wire, "nosuchapp", "", nil); got != error(wire) {
		t.Fatalf("expected the wire error untouched, got %v", got)
	}
}

func TestCancelRejectionIsLeftAloneWhenTheRequestedSourceMissesTheRow(t *testing.T) {
	// A --source the row disagrees with earns the same 404 by itself, so
	// "it most likely finished" would talk past the actual mistake.
	wire := &APIError{
		StatusCode: http.StatusNotFound,
		Message:    "App not found or current state does not allow operation",
	}
	row := &installedAppRow{State: "installing", Source: "market-local"}
	if got := explainCancelRejection(wire, "jellyfin", "appstore", row); got != error(wire) {
		t.Fatalf("expected the wire error untouched, got %v", got)
	}
}

func TestCancelRejectionStillExplainsWhenTheRowCarriesNoSource(t *testing.T) {
	// Nothing to disagree with: an older row without a source cannot rule the
	// settled-state reading out, and the hint ends in `market status` anyway.
	wire := &APIError{
		StatusCode: http.StatusNotFound,
		Message:    "App not found or current state does not allow operation",
	}
	got := explainCancelRejection(wire, "jellyfin", "appstore", &installedAppRow{State: "running"})
	if !strings.Contains(got.Error(), "was in state 'running'") {
		t.Fatalf("expected the settled-state hint, got %v", got)
	}
}

func TestNonNotFoundCancelFailuresAreLeftAlone(t *testing.T) {
	// A 500 is not the ambiguous sentence this helper exists to disambiguate.
	wire := &APIError{StatusCode: http.StatusInternalServerError, Message: "boom"}
	if got := explainCancelRejection(wire, "jellyfin", "", &installedAppRow{State: "running"}); got != error(wire) {
		t.Fatalf("expected the wire error untouched, got %v", got)
	}

	plain := errors.New("connection refused")
	if got := explainCancelRejection(plain, "jellyfin", "", &installedAppRow{State: "running"}); got != plain {
		t.Fatalf("expected the transport error untouched, got %v", got)
	}
}

// The type exists so callers can branch on status without matching prose; the
// prose has to stay identical anyway, because it is what users already see.
func TestAPIErrorReadsAsItDidBeforeTheTypeExisted(t *testing.T) {
	err := &APIError{StatusCode: http.StatusNotFound, Message: "Chart not found"}
	if want := "API error (HTTP 404): Chart not found"; err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}
