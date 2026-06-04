package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/beclab/Olares/cli/pkg/olaresclient"
)

// VersionSuspectError lets a command's operation tell the dispatch aspect "this
// failure might be caused by talking to a backend whose version differs from
// the one I dispatched on". The aspect reacts by re-detecting the backend
// version and, if it changed, re-dispatching and retrying the operation once
// (the self-heal path, analogous to the token refresh-and-retry in
// refreshingTransport). Implement it on a typed error, or use
// MarkVersionSuspect to wrap an existing error.
type VersionSuspectError interface {
	error
	VersionSuspect() bool
}

type versionSuspectError struct{ err error }

func (e *versionSuspectError) Error() string        { return e.err.Error() }
func (e *versionSuspectError) Unwrap() error        { return e.err }
func (e *versionSuspectError) VersionSuspect() bool { return true }

// MarkVersionSuspect wraps err so the dispatch aspect treats it as a possible
// version mismatch and attempts a refresh-and-retry. Returns nil for a nil err.
func MarkVersionSuspect(err error) error {
	if err == nil {
		return nil
	}
	return &versionSuspectError{err: err}
}

// isVersionSuspect reports whether err signals a likely version mismatch worth
// a refresh-and-retry: either it implements VersionSuspectError, or it carries
// an HTTP status the backend uses when a route/shape is unknown (404 Not Found,
// 501 Not Implemented). Auth failures are intentionally excluded — those are
// the refreshingTransport's domain, not a version problem.
func isVersionSuspect(err error) bool {
	if err == nil {
		return false
	}
	var vs VersionSuspectError
	if errors.As(err, &vs) && vs.VersionSuspect() {
		return true
	}
	var he interface{ HTTPStatus() int }
	if errors.As(err, &he) {
		switch he.HTTPStatus() {
		case http.StatusNotFound, http.StatusNotImplemented:
			return true
		}
	}
	return false
}

// WithOlaresClient is the version-dispatch "aspect" for multi-version commands.
// Only commands wired through this entry point get version-aware behavior; all
// other commands are untouched.
//
// It resolves the backend version, builds the matching olaresclient.OlaresClient
// (floor-selected, see olaresclient.GetClient), and runs op with it. Two
// cross-cutting behaviors wrap op:
//
//	A. Stale-version self-heal: if op fails with a version-suspect error
//	   (see isVersionSuspect), the backend version is re-detected; if it
//	   changed, a fresh client is built and op is retried exactly once.
//	B. Capability gate: if op fails with *olaresclient.ErrUnsupportedVersion,
//	   the error is surfaced verbatim (its Error() renders the "requires
//	   Olares >= X" hint) and is NOT retried.
func (f *Factory) WithOlaresClient(ctx context.Context, op func(c olaresclient.OlaresClient) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	version, err := f.OlaresBackendVersion(ctx)
	if err != nil {
		return err
	}
	client, err := olaresclient.GetClient(version)
	if err != nil {
		return err
	}

	runErr := op(client)
	if runErr == nil {
		return nil
	}

	// B. Capability gate — never retried.
	var unsupported *olaresclient.ErrUnsupportedVersion
	if errors.As(runErr, &unsupported) {
		return runErr
	}

	// A. Stale-version self-heal — re-detect, re-dispatch, retry once.
	if !isVersionSuspect(runErr) {
		return runErr
	}
	newVersion, changed, refreshErr := f.RefreshOlaresBackendVersion(ctx)
	if refreshErr != nil || !changed {
		// Couldn't refresh, or the backend version is unchanged — the
		// failure is not a version mismatch we can recover from.
		return runErr
	}
	fmt.Fprintf(os.Stderr, "notice: backend version changed to %s; retrying with the matching client\n", newVersion)
	newClient, err := olaresclient.GetClient(newVersion)
	if err != nil {
		return runErr
	}
	if retryErr := op(newClient); retryErr != nil {
		return retryErr
	}
	return nil
}
