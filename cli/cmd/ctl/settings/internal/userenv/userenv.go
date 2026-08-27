// Package userenv is the client for user-service's env vectors
// (/api/env/systemenvs, /api/env/userenvs), shared by `settings advanced
// env` and `settings appearance theme`.
//
// The PUT upserts exactly the entries in the body and leaves every other
// one untouched (app-service's batchUpdateUserEnvs), so a write sends
// only the variables it means to change. Sending the whole vector back
// is not just redundant: the handler rejects the entire batch with a 400
// if any entry in it is non-editable, so resending unrelated variables
// turns an unrelated read-only entry into a failed write.
package userenv

import (
	"context"
	"fmt"
	"sort"

	"github.com/beclab/Olares/cli/pkg/bflenvelope"
)

// The two vectors user-service exposes.
const (
	SystemEnvsPath = "/api/env/systemenvs"
	UserEnvsPath   = "/api/env/userenvs"
)

// ThemeEnvName is the per-user theme preference. build/user-env.yaml
// declares this UserEnv the single source of truth for the theme, read
// and written by both the env settings page and BFL's config-system
// endpoint.
const ThemeEnvName = "OLARES_USER_THEME"

type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

// Entry mirrors apps/.../constant/index.ts BaseEnv. Callers render a
// subset in tables and marshal the whole struct for --output json, so
// the optional fields stay omitempty to keep that output identical to
// what the upstream sent.
type Entry struct {
	EnvName     string `json:"envName"`
	Value       string `json:"value,omitempty"`
	Default     string `json:"default,omitempty"`
	Editable    *bool  `json:"editable,omitempty"`
	Type        string `json:"type,omitempty"`
	Required    *bool  `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

// Pair is one element of the PUT body — matches UpdateEnvItem in
// constant/index.ts.
type Pair struct {
	EnvName string `json:"envName"`
	Value   string `json:"value"`
}

// List reads the current vector at path.
func List(ctx context.Context, d Doer, path string) ([]Entry, error) {
	var entries []Entry
	if err := doEnvelope(ctx, d, "GET", path, nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SetValues writes the named variables. The upstream rejects the whole
// batch when any entry in it is flagged editable: false, or when a
// required entry would end up empty.
func SetValues(ctx context.Context, d Doer, path string, updates map[string]string) error {
	if len(updates) == 0 {
		return fmt.Errorf("no environment variable updates supplied")
	}
	return doEnvelope(ctx, d, "PUT", path, Body(updates), nil)
}

// Body renders updates as the PUT payload, sorted by name so a
// multi-variable write does not depend on Go's map iteration order.
func Body(updates map[string]string) []Pair {
	names := make([]string, 0, len(updates))
	for name := range updates {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]Pair, 0, len(names))
	for _, name := range names {
		out = append(out, Pair{EnvName: name, Value: updates[name]})
	}
	return out
}

func doEnvelope(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env bflenvelope.Envelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	return bflenvelope.Data(method, path, env, out)
}
