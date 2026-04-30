// Package mkdir implements the wire side of `olares-cli files mkdir`,
// hitting the per-user files-backend's "create directory" endpoint:
//
//	POST /api/resources/<fileType>/<extend><subPath>/
//
// The trailing slash on the URL is what the backend uses to
// distinguish "create directory" from "create empty file"
// (postCreateFile in apps/.../api/files/v2/common/utils.ts does the
// same thing — `isDir ? '/' : ''`). Body is empty.
//
// Wire shape parity across namespaces: the LarePass web app's
// per-driver `createDir` helpers (drive/utils.ts, sync/utils.ts,
// awss3/utils.ts, dropbox/utils.ts, google/utils.ts, tencent/utils.ts,
// cache/utils.ts, external/utils.ts) all funnel through the same
// `commonUrlPrefix('resources') + <fileType>/<extend>/...` POST, so
// this verb is supported uniformly across drive / sync / cache /
// external / awss3 / google / dropbox / tencent. The auto-rename
// quirk (see the comment on Client.Mkdir for the gory details) is
// also uniform.
//
// `-p` / `--parents` semantics live in the cobra layer (cmd/ctl/files/
// mkdir.go) on top of two primitives this package exports:
//
//   - Plan / PlanRecursive: input validation + Op materialisation,
//     no I/O.
//   - Client.Mkdir: the actual POST.
//   - Client.Exists: parent-directory listing + basename lookup, used
//     by `-p` mode to skip already-existing intermediate directories
//     (the auto-rename behavior would otherwise turn `mkdir -p A/B/C`
//     into "A (1)/B/C" if A already exists).
//
// Why a per-package Exists rather than reaching into `download.Stat`:
// each verb package under cli/internal/files/ is self-contained (cp /
// rm / rename / share / repos / download / upload — none of them
// cross-import each other). Keeping that invariant means a future
// move/refactor of any one package doesn't ripple through the rest.
// The Exists helper here is intentionally minimal — it only decodes
// (Name, IsDir) from the parent listing — and accepts both the Drive
// `items` envelope and the cloud-drive `data` envelope.
package mkdir

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Target is one user-supplied path, normalized so the planner has a
// single canonical shape to validate. The cobra layer's path parser
// produces this; this package intentionally doesn't import
// FrontendPath so it stays free of cobra/cmdutil deps — same
// convention as the cp / rm / rename / share packages.
type Target struct {
	// FileType + Extend + SubPath together form the wire path
	// (joined with '/' and percent-encoded per segment): e.g.
	// ("drive", "Home", "/Documents/Backups") →
	// /api/resources/drive/Home/Documents/Backups/.
	FileType string
	Extend   string
	// SubPath always starts with '/'. Whether or not it ends with
	// '/' is irrelevant for mkdir — the verb always means "this
	// path is a directory" — but we preserve the user's input
	// shape in DisplayPath so the log line reads naturally.
	SubPath string
}

// Op is one POST /api/resources/<...>/ call, fully resolved. Endpoint
// is the URL relative to BaseURL (already percent-encoded), so the
// HTTP call site doesn't re-encode. DisplayPath is the human-readable
// 3-segment frontend form (e.g. `drive/Home/Documents/Backups/`)
// surfaced in log lines and error messages.
type Op struct {
	Endpoint    string
	DisplayPath string
}

// Client is the per-FilesURL handle for mkdir + existence checks.
// HTTPClient is expected to be the factory-provided client whose
// refreshingTransport injects `X-Authorization` (NOT `Authorization:
// Bearer`, see pkg/cmdutil/factory.go for why) and refreshes on
// 401/403 transparently.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// HTTPError carries the status + truncated body of a non-2xx response
// so the cobra layer can branch on the status code (401 / 403 / 404 /
// 409) with friendly CTAs. Same shape as the per-package errors in
// cp / rm / rename / download for a uniform error contract.
type HTTPError struct {
	Status int
	Body   string
	URL    string
	Method string
}

func (e *HTTPError) Error() string {
	body := e.Body
	if len(body) > 500 {
		body = body[:500] + "...(truncated)"
	}
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.URL, e.Status, body)
}

// Plan validates the inputs and returns a single Op for one mkdir
// call. Validation rules:
//
//   - FileType / Extend must be non-empty (defense in depth — the
//     cobra-layer parser shouldn't let these through, but a typed
//     error here beats a silent /api/resources//Home/... URL).
//   - SubPath, after trimming '/', must be non-empty — refusing to
//     "create" the volume root mirrors the rest of the CLI's safety
//     policy (rm / rename / cp all refuse the root).
//   - No segment may be empty, '.', or '..' — these are
//     path-traversal grenades that the server might or might not
//     guard against; in either case they are not what the user
//     meant.
func Plan(t Target) (Op, error) {
	if t.FileType == "" || t.Extend == "" {
		return Op{}, fmt.Errorf("mkdir: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	clean := strings.Trim(t.SubPath, "/")
	if clean == "" {
		return Op{}, fmt.Errorf(
			"refusing to mkdir the root of %s/%s; pick a subdirectory name (e.g. %s/%s/NewFolder)",
			t.FileType, t.Extend, t.FileType, t.Extend)
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return Op{}, fmt.Errorf(
				"mkdir: path segment %q is invalid (empty / '.' / '..' are not real directory names): %s",
				seg, t.FileType+"/"+t.Extend+t.SubPath)
		}
	}

	return Op{
		Endpoint:    buildEndpoint(t.FileType, t.Extend, clean),
		DisplayPath: t.FileType + "/" + t.Extend + "/" + clean + "/",
	}, nil
}

// PlanRecursive splits the requested target into one Op per missing
// path segment, in left-to-right order, so the cobra layer's `-p`
// mode can step through them. Each Op covers exactly one prefix — so
// for `drive/Home/A/B/C` you get three Ops creating `A`, then `A/B`,
// then `A/B/C`.
//
// The cobra layer is expected to interleave Client.Exists calls so it
// can skip Ops whose prefix already exists; without that, the
// auto-rename quirk (POST `/A/` when `A` exists creates `A (1)`)
// would silently produce a parallel directory tree. Plan does NOT do
// the existence check here because we want network I/O to remain
// confined to Client.* methods — keeps the planner deterministic and
// unit-testable.
//
// PlanRecursive shares Plan's validation rules, applied to the full
// SubPath up front so a malformed input fails fast (we don't want to
// create the first two of three intermediate dirs and then refuse the
// third).
func PlanRecursive(t Target) ([]Op, error) {
	if _, err := Plan(t); err != nil {
		return nil, err
	}
	clean := strings.Trim(t.SubPath, "/")
	segs := strings.Split(clean, "/")
	out := make([]Op, 0, len(segs))
	for i := range segs {
		prefix := strings.Join(segs[:i+1], "/")
		out = append(out, Op{
			Endpoint:    buildEndpoint(t.FileType, t.Extend, prefix),
			DisplayPath: t.FileType + "/" + t.Extend + "/" + prefix + "/",
		})
	}
	return out, nil
}

// buildEndpoint returns the percent-encoded `/api/resources/...` URL
// path with a trailing slash (the backend's "this is a directory"
// signal). `cleanSub` must be the SubPath already trimmed of leading
// and trailing '/'. We percent-encode the full plain form once
// (rather than per segment + join) so embedded '/' separators stay
// unescaped — same convention as encodepath.EncodeURL elsewhere in
// the CLI.
func buildEndpoint(fileType, extend, cleanSub string) string {
	plain := fileType + "/" + extend + "/" + cleanSub
	return "/api/resources/" + encodepath.EncodeURL(plain) + "/"
}

// Mkdir POSTs an empty body against op.Endpoint to create the
// directory. A 409 from the server is treated as "already exists" and
// returns nil — that lets `mkdir -p` paper over a race where another
// client created the same intermediate dir between our Exists probe
// and our POST.
//
// IMPORTANT (collision behavior): on the current files-backend the
// 409 path is rare. For drive / sync / cache / external the server
// AUTO-RENAMES on collision (POST creates "Foo (1)" if "Foo" already
// exists) instead of returning 409. So `mkdir foo/` where `foo`
// already exists produces a SECOND directory with a suffix, NOT an
// error. The cobra layer's `-p` mode side-steps this by checking
// Exists before issuing the POST; non-`-p` mkdir surfaces this
// behavior in its post-run summary so the user can ls and confirm.
//
// Errors:
//   - non-2xx response → *HTTPError so the cobra layer can branch on
//     401 / 403 / 404 with friendly CTAs.
//   - request build / network failures bubble up verbatim.
func (c *Client) Mkdir(ctx context.Context, op Op) error {
	if op.Endpoint == "" {
		return errors.New("Mkdir: empty Endpoint (Plan should have rejected this)")
	}
	endpoint := c.BaseURL + op.Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	// Accept JSON so a server that returns an envelope on 4xx gives
	// us the structured error message instead of an HTML page.
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		// Already exists — exactly what we'd want for an
		// idempotent ensure-dir step. The current backend rarely
		// takes this branch (auto-rename instead), but we keep it
		// for defense in depth.
		return nil
	}
	if resp.StatusCode/100 != 2 {
		return &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodPost,
		}
	}
	return nil
}

// Exists reports whether a remote path already exists by listing its
// parent and looking up the basename. Returns:
//
//   - found: true if the basename is in the parent listing.
//   - isDir: true if the found entry is a directory; meaningful only
//     when found is true.
//   - error: parent-listing failures bubble up; a 404 on the parent
//     listing surfaces as found=false (the parent doesn't exist
//     either, which is also "this path doesn't exist").
//
// The plain form is `<fileType>/<extend>/<sub>` (no leading slash,
// trailing slash optional — both are tolerated). Used by the cobra
// layer's `-p` mode to decide whether to skip a segment.
//
// Why parent-listing rather than a direct GET on the path: the
// per-user files-backend's GetResources handler reads file content
// into the response on a single-resource GET (see the comment on
// download/stat.go for the gory details), which can produce HTTP 500
// for many real files. Listing the parent and matching the basename
// is the only strategy the wire reliably supports — and it's what
// the LarePass web app uses for navigation.
func (c *Client) Exists(ctx context.Context, plainPath string) (found bool, isDir bool, err error) {
	clean := strings.Trim(plainPath, "/")
	if clean == "" {
		// Volume / tree root — always present, always a directory.
		return true, true, nil
	}
	segs := strings.Split(clean, "/")
	if len(segs) <= 2 {
		// `<fileType>/<extend>` — the volume / extend root is a
		// directory by definition; no listing needed.
		return true, true, nil
	}
	leaf := segs[len(segs)-1]
	parentPlain := strings.Join(segs[:len(segs)-1], "/") + "/"
	endpoint := c.BaseURL + "/api/resources/" + encodepath.EncodeURL(parentPlain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, false, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, false, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		// Parent doesn't exist either — `-p` mode will create both
		// in turn.
		return false, false, nil
	}
	if resp.StatusCode/100 != 2 {
		return false, false, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}

	// Two envelope shapes are accepted: Drive/Sync `items` and the
	// cloud-drive `data` array. We only need (Name, IsDir) — `mode`
	// / `modified` empty-string quirks on cloud entries don't
	// matter for the existence check, so we DON'T import the
	// fancier listing decoder from cmd/ctl/files/ls.go (that would
	// pull cobra into this package).
	var env struct {
		Items []listItem `json:"items"`
		Data  []listItem `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return false, false, fmt.Errorf("decode listing for %s: %w", parentPlain, err)
	}
	src := env.Items
	if len(src) == 0 && len(env.Data) > 0 {
		src = env.Data
	}
	for _, it := range src {
		if it.Name == leaf {
			return true, it.IsDir, nil
		}
	}
	return false, false, nil
}

// listItem is the deliberately-narrow per-entry projection Exists
// needs. `IsDir` and `Name` are populated identically across the
// Drive `items` and cloud-drive `data` envelopes — we don't decode
// `size` / `mode` / `modified` here because Exists doesn't care.
type listItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

// IsHTTPStatus is a convenience predicate the cobra layer uses to
// branch on common 4xx codes. Same shape as cp.IsHTTPStatus / etc.;
// duplicated here to keep this package self-contained.
func IsHTTPStatus(err error, status int) bool {
	var hErr *HTTPError
	return errors.As(err, &hErr) && hErr.Status == status
}
