// Package whoami centralizes the "who am I on this Olares?" round-trip.
// It is shared between three command surfaces that all answer the same
// question with the same backend hit:
//
//	olares-cli profile whoami
//	olares-cli settings users me
//	olares-cli settings me whoami
//
// The eager refresh on `profile login` / `profile import` also flows through
// here (FetchAndCache, with the "best-effort, non-fatal" semantics
// documented on that helper).
//
// The package is deliberately cobra-free: command files in
// cli/cmd/ctl/profile and cli/cmd/ctl/settings/{users,me} import this
// package and wrap it in a thin RunE; the heavy lifting (HTTP, decode,
// drift detection, atomic config write) lives here once.
package whoami

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// Endpoint is the user-service path that proxies to BFL's
// /bfl/backend/v1/user-info handler. Same endpoint the SPA hits — see
// apps/packages/app/src/stores/settings/admin.ts (`/api/backend/v1/user-info`)
// and user-service/src/bfl/backend.controller.ts (`@Get('/v1/user-info')`).
const Endpoint = "/api/backend/v1/user-info"

// Wire role values, kept identical to BFL's
// framework/bfl/pkg/constants/constants.go (RoleOwner / RoleAdmin /
// RoleNormal). We re-declare them here rather than depend on the BFL Go
// module because the CLI doesn't pull in that module today and shouldn't
// start now just for three string constants.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleNormal = "normal"
)

// roleRank gives "is X at least Y?" a simple numeric comparison for the
// soft-preflight gate. The owner-only tier (3) and admin-or-owner tier (2)
// are the two non-trivial ones; every other settings verb is roleNormal
// (1) — every authenticated user passes.
//
// Empty / unknown role ranks 0 — preflight callers MUST treat 0 as "skip
// the local check, let the server be the source of truth", to keep
// pre-existing profiles working without forcing a re-login.
var roleRank = map[string]int{
	RoleNormal: 1,
	RoleAdmin:  2,
	RoleOwner:  3,
}

// FriendlyLabel returns the user-facing label for a wire role: the SPA and
// docs call "normal" → "User", but everywhere else uses the wire spelling.
// Used by printers; preflight code MUST keep using the wire value.
func FriendlyLabel(wire string) string {
	switch wire {
	case RoleOwner:
		return "Owner"
	case RoleAdmin:
		return "Admin"
	case RoleNormal:
		return "User"
	case "":
		return "(unknown)"
	default:
		// Forward-compat: future BFL roles stay readable instead of
		// silently disappearing. Lower-case + Title-case to match the
		// shape of the three known labels.
		if len(wire) == 0 {
			return wire
		}
		return strings.ToUpper(wire[:1]) + strings.ToLower(wire[1:])
	}
}

// Rank returns the numeric tier for `role`, or 0 for the unknown / empty
// case. Used by Settings preflight code: `Rank(cached) >= Rank(required)`
// passes; `0` (unknown) is treated as pass to avoid blocking on a stale or
// missing cache (server is the authority).
func Rank(role string) int {
	if r, ok := roleRank[role]; ok {
		return r
	}
	return 0
}

// Info is the in-memory model of the BFL UserInfo response. Field names
// keep the wire JSON tags so callers can `--output json` straight from the
// struct without an extra mapping layer. AccessLevel is omitempty because
// BFL only sets it when the user-launcher annotation is present.
//
// Unmapped wire fields (zone, is_ephemeral, terminusName, created_user,
// wizard_complete) are intentionally not modeled here yet — Phase 0b only
// needs name + owner_role for the cache. Add them as later phases find
// uses; doing it now would invite "decoded but never read" lint pressure.
type Info struct {
	Name      string `json:"name"`
	OwnerRole string `json:"owner_role"`
}

// envelope decodes the BFL wrapper that user-service forwards verbatim
// (see framework/bfl/pkg/api/response/response.go's Response). The CLI does
// NOT have a global axios-style interceptor like the SPA's
// apps/packages/app/src/boot/axios.ts (which strips data.data into the
// returned value), so we own the unwrap here.
type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Info   `json:"data"`
}

// Doer is the minimal HTTP surface we need from settings.SettingsClient
// (and any future caller). Defining it locally keeps the import graph tidy:
// the whoami package depends on neither cmdutil nor settings, just on
// cliconfig + net/http.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

// Result is what callers get back from FetchAndCache: the freshly-decoded
// Info plus a "did this overwrite a different cached role?" flag and the
// previous wire value (when one existed).
//
// Changed=true only fires for actual transitions (admin→owner,
// owner→admin, ...). First-time writes (empty → role) also report
// Changed=true so first-login UX gets the same "your role is X" line as
// genuine changes.
type Result struct {
	Info             Info
	Changed          bool
	PreviousRole     string
	WroteToCache     bool  // false when caller passed cfg=nil (profile not in config.json)
	RefreshedAt      int64 // Unix-second timestamp written to cache
	AlreadyMatchedAt int64 // populated only when Changed=false; the previous WhoamiRefreshedAt
}

// FetchAndCache hits Endpoint with `client`, decodes the BFL envelope, and
// (when cfg is non-nil) atomically updates the matching profile's
// OwnerRole + WhoamiRefreshedAt fields.
//
// olaresID is required so the cache write targets the right profile —
// callers that already have a ResolvedProfile usually pass rp.OlaresID
// here. Pre-existing profiles that don't match olaresID return an error
// from cliconfig.SetOwnerRole; callers typically demote that to a warning
// (the network round-trip succeeded; only the cache write failed).
//
// `now` is injected for testability; pass time.Now in production.
//
// On HTTP / decode failure FetchAndCache returns the error untouched —
// the eager-fetch caller (profile login / import) is expected to wrap it
// in a non-fatal warning, while the explicit `whoami` caller surfaces it
// as a regular error.
func FetchAndCache(
	ctx context.Context,
	client Doer,
	cfg *cliconfig.MultiProfileConfig,
	olaresID string,
	now func() time.Time,
) (*Result, error) {
	if client == nil {
		return nil, errors.New("whoami: nil http client")
	}
	if olaresID == "" {
		return nil, errors.New("whoami: empty olaresID")
	}
	if now == nil {
		now = time.Now
	}

	var env envelope
	if err := client.DoJSON(ctx, http.MethodGet, Endpoint, nil, &env); err != nil {
		return nil, err
	}
	// BFL wraps every response in {code,message,data}; user-service forwards
	// it verbatim. Non-zero code is a server-side rejection that didn't
	// surface as an HTTP non-2xx (rare, but BFL does it for some validation
	// errors) — surface it the same way the SPA does in
	// apps/packages/app/src/boot/axios.ts:159 (`if (data.code != 0) throw`).
	if env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("server returned code=%d", env.Code)
		}
		return nil, fmt.Errorf("whoami: %s", msg)
	}

	res := &Result{Info: env.Data, RefreshedAt: now().Unix()}
	if cfg == nil {
		// In-memory only — used when the resolved profile came from the
		// EnvProvider (no on-disk profile to update) or test scaffolds.
		return res, nil
	}

	target := cfg.FindByOlaresID(olaresID)
	if target == nil {
		return nil, fmt.Errorf("whoami: profile %q not found in config", olaresID)
	}
	res.PreviousRole = target.OwnerRole
	res.AlreadyMatchedAt = target.WhoamiRefreshedAt

	changed, err := cfg.SetOwnerRole(olaresID, env.Data.OwnerRole, res.RefreshedAt)
	if err != nil {
		return nil, err
	}
	// SetOwnerRole already encodes "non-empty new role and prev != new" as
	// changed, which covers both first-time writes (empty -> X) and genuine
	// transitions (X -> Y). Empty -> empty and X -> X return false. We rely
	// on that semantics directly so we don't second-guess the helper.
	res.Changed = changed
	res.WroteToCache = true
	return res, nil
}
