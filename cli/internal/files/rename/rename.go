// Package rename implements the wire side of `olares-cli files rename`,
// hitting the per-user files-backend's in-place rename endpoint:
//
//	PATCH /api/resources/<fileType>/<extend><subPath>[/]?destination=<newName>
//
// This is a SEPARATE wire surface from `cp` / `mv` (which route through
// PATCH /api/paste/<node>/, async task queue). Rename is:
//
//   - Same parent — only the basename changes.
//   - Synchronous — the response is the final state, no task_id polling.
//   - No <node> URL segment; works against drive / sync / cloud / external
//     uniformly because /api/resources is the catch-all per-resource path.
//   - Body is empty; the new name lives in the `destination` query param
//     and is encodeURIComponent-shaped (JS-compatible — see internal/files/encodepath
//     for why we don't reuse net/url here).
//
// Source of truth on the wire shape is the LarePass web app's
// renameFileItem helper in
// apps/packages/app/src/api/files/v2/common/utils.ts L179-L198: the
// PATCH URL is /api/resources/<fileType>/<extend><encUrl(oPath)>[/],
// and the only query param the standard rename modal sends is
// `destination` (encodeURIComponent of the new basename).
package rename

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Target is one user-supplied source path, normalized so the planner
// has a single canonical shape to validate. Construct via the cobra
// layer's path parser; the planner intentionally doesn't know about
// FrontendPath so this package stays free of cobra/cmdutil deps —
// same convention as the cp / rm packages.
type Target struct {
	// FileType + Extend + SubPath together form the wire path
	// (joined with '/' and percent-encoded per segment): e.g.
	// ("drive", "Home", "/Documents/foo.pdf") →
	// /api/resources/drive/Home/Documents/foo.pdf.
	FileType string
	Extend   string
	// SubPath always starts with '/'. A trailing '/' (preserved
	// from user input) is the directory marker that the backend
	// uses to route the rename through the directory or file
	// handler — keep it explicit so we don't have to re-stat just
	// to figure out which kind of resource we're renaming.
	SubPath string
	// IsDirIntent: did the user signal this is a directory (trailing
	// '/' on the path)? Mirrored to the URL's trailing '/' on the
	// wire and informs nothing else here — rename is non-recursive
	// from the client's POV (the server moves whatever sits at the
	// path atomically).
	IsDirIntent bool
}

// Op is one PATCH /api/resources/.../?destination=... call, fully
// resolved. Endpoint is the full URL relative to BaseURL (already
// percent-encoded for both the path and the query value), so the
// http call site can build the request without re-encoding anything.
//
// DisplaySrc / DisplayDst are the human-readable forms used for log
// lines and error messages; we keep them separate from Endpoint so
// the wire shape is allowed to evolve (extra query params, version
// segment) without breaking the user-visible output.
type Op struct {
	Endpoint   string
	DisplaySrc string
	DisplayDst string
	IsDir      bool
}

// Client is the per-FilesURL handle for rename calls. HTTPClient is
// expected to be the factory-provided client whose refreshingTransport
// injects `X-Authorization` (not `Authorization: Bearer`, see
// pkg/cmdutil/factory.go for why) and refreshes on 401/403.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// HTTPError carries the status + truncated body of a non-2xx response
// so the cobra layer can branch on the status code (e.g. to give a
// friendly "not found" or auth-issue CTA). Same shape as the
// per-package HTTP errors in upload / rm / cp / download for a
// uniform error contract.
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

// Plan validates the rename inputs and produces an Op that the cobra
// layer can hand to Client.Rename. Validation rules:
//
//   - newName must not be empty after trimming whitespace.
//   - newName must not contain '/' or '\\' — rename is in-place by
//     construction; cross-directory moves are `cp` / `mv` territory.
//   - newName must not be exactly "." or ".." — these are
//     path-traversal grenades that the server might or might not
//     guard against, and in either case they're never what the user
//     meant.
//   - The source must not be the volume root (SubPath == "/") — same
//     refusal pattern as rm and cp; "rename my whole Drive/Home" is
//     not a meaningful operation.
//   - newName equal to the source's current basename is rejected —
//     it's a no-op and almost always a typo.
//
// The Endpoint is built with internal/files/encodepath: per-segment
// JS-shaped percent-encoding for the path, encodeURIComponent shape
// for the destination query value. This stays aligned with the rest
// of the CLI (rm / upload / download / cat) so a path that lists
// fine via `files ls` also renames via the same encoded bytes.
func Plan(t Target, newName string) (Op, error) {
	if t.FileType == "" || t.Extend == "" {
		// Defense in depth — the cobra-layer parser shouldn't let
		// these through, but a typed error here beats a silent
		// "/api/resources//Home/..." URL.
		return Op{}, fmt.Errorf("rename: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	if strings.Trim(t.SubPath, "/") == "" {
		return Op{}, fmt.Errorf(
			"refusing to rename the root of %s/%s",
			t.FileType, t.Extend)
	}

	clean := strings.TrimSpace(newName)
	if clean == "" {
		return Op{}, errors.New("rename: new name is empty")
	}
	if strings.ContainsAny(clean, "/\\") {
		// Single message for both '/' and '\\': the backend rejects
		// backslash anywhere in a wire path (the same `code: -1`
		// failure mode `cp` / `mv` surface), and forward slashes
		// would silently turn rename into "move to a subdirectory"
		// which is not what this verb does.
		return Op{}, fmt.Errorf(
			"rename: new name %q must not contain '/' or '\\\\' (rename is in-place; use `files mv` for cross-directory moves)",
			clean)
	}
	if clean == "." || clean == ".." {
		return Op{}, fmt.Errorf("rename: new name %q is a path-traversal segment, not a real name", clean)
	}

	currentBase := lastSegment(t.SubPath)
	if currentBase == "" {
		// Same root-rejection as above, except we couldn't tell
		// from the SubPath alone — guard so we don't produce a
		// wire URL with a missing basename.
		return Op{}, fmt.Errorf(
			"refusing to rename %s/%s: cannot derive a current basename from %q",
			t.FileType, t.Extend, t.SubPath)
	}
	if clean == currentBase {
		return Op{}, fmt.Errorf(
			"rename: new name %q is the same as the current basename; nothing to do",
			clean)
	}

	// Build the wire URL. plain is the unencoded human-readable form
	// the rest of the CLI uses (e.g. "drive/Home/Documents/foo.pdf");
	// EncodeURL handles per-segment percent-encoding while preserving
	// '/' separators and trailing-slash directory hints.
	plain := t.FileType + "/" + t.Extend + t.SubPath
	if t.IsDirIntent && !strings.HasSuffix(plain, "/") {
		plain += "/"
	}
	encPath := encodepath.EncodeURL(plain)
	encName := encodepath.EncodeURIComponent(clean)
	endpoint := "/api/resources/" + encPath + "?destination=" + encName

	// Compute the human-readable display destination too — same path
	// as Source, but with the basename replaced. Useful in the
	// command summary so the user sees the full new path, not just
	// the bare new name.
	displaySrc := t.FileType + "/" + t.Extend + t.SubPath
	if t.IsDirIntent && !strings.HasSuffix(displaySrc, "/") {
		displaySrc += "/"
	}
	displayDst := replaceLastSegment(displaySrc, clean, t.IsDirIntent)

	return Op{
		Endpoint:   endpoint,
		DisplaySrc: displaySrc,
		DisplayDst: displayDst,
		IsDir:      t.IsDirIntent,
	}, nil
}

// Rename sends one PATCH against the planned endpoint. Body is empty
// (the destination lives in the query string) and Content-Type is
// not set — the backend's resource handler treats a missing body as
// "no per-attribute update", which is exactly what rename wants.
//
// Errors:
//   - non-2xx response → *HTTPError so the cobra layer can branch on
//     401 / 403 / 404 / 409 with friendly CTAs.
//   - request build / network failures bubble up verbatim.
func (c *Client) Rename(ctx context.Context, op Op) error {
	if op.Endpoint == "" {
		return errors.New("Rename: empty Endpoint (Plan should have rejected this)")
	}

	endpoint := c.BaseURL + op.Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, nil)
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
	if resp.StatusCode/100 != 2 {
		return &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodPatch,
		}
	}
	return nil
}

// lastSegment returns the basename of a slash-separated subpath,
// ignoring leading and trailing '/'. Same shape as the cp package's
// helper of the same name; duplicated here to keep the rename
// package free of a dependency on cp (cp pulls in JSON / paste
// machinery this package doesn't need).
func lastSegment(sub string) string {
	s := strings.Trim(sub, "/")
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// replaceLastSegment returns the human-readable display path with the
// last basename swapped for newBase. Used only for log lines, NOT for
// the wire URL — the wire URL is built from the encoded src path and
// the encoded query value, both of which are produced separately in
// Plan.
//
// Preserves a trailing '/' when isDir is true so the displayed
// destination keeps its directory marker (matching what `files ls`
// would print).
func replaceLastSegment(plain, newBase string, isDir bool) string {
	trimmed := strings.TrimRight(plain, "/")
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 {
		// Defensive — Plan rejects volume-root sources, so the
		// path always has at least one '/' separator.
		return trimmed + "/" + newBase
	}
	out := trimmed[:idx+1] + newBase
	if isDir {
		out += "/"
	}
	return out
}
