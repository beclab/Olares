// Package edge adapts the `settings` per-user edge transport to the
// olaresclient.Doer contract used by the version-compatibility layer.
//
// The version clients in pkg/olaresclient shape only the request line
// (method / path / body) and expect Doer.Do to return the unwrapped `data`
// of the BFL `{code, message, data}` envelope. Market commands satisfy that
// over the app-store transport; the `settings` areas (gpu, network overlay,
// ...) talk to user-service on the profile's DesktopURL via
// whoami.NewHTTPClient. This package provides one small adapter both reuse so
// the envelope-unwrap logic lives in exactly one place.
package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// jsonDoer is the minimal HTTP surface this adapter wraps (one
// JSON-in/JSON-out method). *whoami.HTTPClient satisfies it.
type jsonDoer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

// Doer implements olaresclient.Doer over the settings edge transport. It is
// satisfied structurally — callers pass *Doer where olaresclient.Doer is
// expected, so this package need not import olaresclient (avoiding a cycle).
type Doer struct {
	inner jsonDoer
}

// New builds a Doer for the active profile, returning the resolved profile so
// callers that need identity (e.g. the overlay status `{user}` path) can read
// it without a second lookup.
func New(ctx context.Context, f *cmdutil.Factory) (*Doer, *credential.ResolvedProfile, error) {
	if f == nil {
		return nil, nil, fmt.Errorf("internal error: settings edge not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return &Doer{inner: whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)}, rp, nil
}

// envelope is the BFL `{code, message, data}` wrapper user-service forwards
// (or wraps via returnSucceed). code 0 or 200 means success.
type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Do sends a JSON request and returns the unwrapped `data` payload. A non
// 0/200 code is surfaced as an error (some BFL validation failures arrive as
// HTTP 200 with a non-zero body code, mirroring the SPA's interceptor).
func (d *Doer) Do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var env envelope
	if err := d.inner.DoJSON(ctx, method, path, body, &env); err != nil {
		return nil, err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return nil, fmt.Errorf("%s %s: upstream returned code %d", method, path, env.Code)
		}
		return nil, fmt.Errorf("%s %s: upstream returned code %d: %s", method, path, env.Code, msg)
	}
	return env.Data, nil
}
