// compress.go: wire side of POST /api/archive/<node>/compress.
//
// The endpoint takes a JSON body describing N sources, one
// destination, the desired format / compression level / volume
// size / symlink policy / collision policy, and returns the task
// id of the queued archive-build task. The actual byte writing
// happens on the per-node task queue (same model as
// /api/paste/<node>/) — the response only tells us "the task is
// queued"; callers that need completion semantics poll via the
// task.go helper.
//
// JSON body shape (mirrors the spec verbatim — keep in sync if
// the backend ever renames a field):
//
//	{
//	  "sources":          ["/drive/Home/folder", "/drive/Home/file.txt"],
//	  "destination":      "/drive/Home/out.zip",
//	  "format":           "zip",
//	  "level":            5,
//	  "volumeSizeMB":     100,
//	  "preserveSymlinks": false,
//	  "conflict":         "rename"
//	}
//
// Path shape on sources / destination: the canonical
// `/<fileType>/<extend>/<sub>` form — same as the `source` /
// `destination` strings in the `cp` package's paste body. The
// caller builds them via BuildWirePath; the package never sees
// the FrontendPath type so it stays free of cobra/cmdutil deps.
package archive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// CompressOptions captures every knob the compress endpoint
// accepts. Construct via the cobra layer's option parser; the
// package's only responsibility is to marshal this into a JSON
// body that matches the spec.
//
// Sentinel values:
//   - Level == -1   → field omitted (backend uses its codec default).
//   - VolumeSizeMB == 0 → field omitted (single-volume archive).
//
// Password is NOT a struct field because the spec ships it via
// the `X-Archive-Password` header, not the body. Pass it as the
// final arg to Compress so the call site documents the security-
// sensitive payload distinctly.
type CompressOptions struct {
	// Sources are already-built wire paths
	// (`/<fileType>/<extend>/<sub>` form, no double slashes, no
	// trailing slash on file sources, trailing slash on dir
	// sources). The cobra layer normalises these before calling.
	Sources []string

	// Destination is the wire path for the archive file the
	// server will produce. Must NOT exist yet (or the conflict
	// policy decides what happens).
	Destination string

	// Format is the validated archive format (one of
	// SupportedFormats). The cobra layer ensures this is set
	// before reaching here.
	Format string

	// Level: 0..9 inclusive, or -1 to omit. ValidateLevel
	// enforces the bounds.
	Level int

	// VolumeSizeMB: split-archive threshold in MiB. 0 omits
	// the field. Only meaningful for zip / 7z; the format
	// gate refuses to send it on other formats.
	VolumeSizeMB int

	// PreserveSymlinks: when true, the archive includes symlinks
	// as-is rather than dereferencing them. False (the default)
	// matches the spec's documented default and the LarePass GUI
	// behavior.
	PreserveSymlinks bool

	// Conflict is the on-collision policy at the destination
	// path. Empty string → ConflictDefault (rename) before send.
	Conflict Conflict

	// Node is the {node} URL segment for /api/archive/<node>/.
	// The cobra layer resolves this via the same cascade `files
	// cp` uses (--node override → dst External/Cache extend →
	// /api/nodes/ default). Empty is rejected up front so the
	// request never leaves with a malformed URL.
	Node string
}

// compressRequestBody is the JSON shape the backend's compress
// handler binds to. We use pointer / omitempty for the optional
// numeric fields so the sentinel "unset" values don't show up on
// the wire as `0` (which the server would interpret as "max
// compression" / "single volume" — both incorrect for the unset
// case).
type compressRequestBody struct {
	Sources          []string `json:"sources"`
	Destination      string   `json:"destination"`
	Format           string   `json:"format"`
	Level            *int     `json:"level,omitempty"`
	VolumeSizeMB     *int     `json:"volumeSizeMB,omitempty"`
	PreserveSymlinks bool     `json:"preserveSymlinks"`
	Conflict         string   `json:"conflict,omitempty"`
}

// compressResponseEnvelope mirrors the spec's success response:
//
//	{ "code": 0, "msg": "success", "task_id": "<id>" }
//
// The server's error responses can carry either a numeric `code`
// distinct from 0 OR a non-2xx HTTP status — we treat the HTTP
// status as the primary signal (httpErrorFromResponse handles
// it) and only branch on `code` for the in-band-success path
// where the wire was 2xx but the payload says "no, actually".
type compressResponseEnvelope struct {
	Code   *int   `json:"code,omitempty"`
	Msg    string `json:"msg,omitempty"`
	TaskID string `json:"task_id,omitempty"`
}

// Compress sends one POST /api/archive/<node>/compress for the
// supplied options and returns the resulting task_id. The actual
// byte writing happens asynchronously on the files-backend's
// task queue — by the time this returns we only know "the
// server has queued the task". Callers that need completion
// semantics call WaitTask afterwards.
//
// Validation runs client-side before the HTTP request:
//
//   - len(Sources) > 0  — refuse empty-batch compress.
//   - Destination != "" — refuse missing target.
//   - Format is one of SupportedFormats.
//   - Level in MinLevel..MaxLevel (or -1 for unset).
//   - VolumeSizeMB only applies to multi-volume formats.
//   - Password (when non-empty) only on passwordable formats.
//
// HTTP errors (auth / network / 4xx / 5xx) surface as *HTTPError;
// the cobra layer's reformatter attaches the standard CTAs.
func (c *Client) Compress(ctx context.Context, opts CompressOptions, password string) (string, error) {
	if err := validateCompressOptions(opts, password); err != nil {
		return "", err
	}

	body := compressRequestBody{
		Sources:          opts.Sources,
		Destination:      opts.Destination,
		Format:           strings.ToLower(opts.Format),
		PreserveSymlinks: opts.PreserveSymlinks,
	}
	if opts.Conflict != "" {
		body.Conflict = string(opts.Conflict)
	}
	if opts.Level >= 0 {
		// Take address of a local — opts.Level is a value, and
		// json.Marshal needs a pointer to distinguish "unset"
		// from "explicit 0".
		lv := opts.Level
		body.Level = &lv
	}
	if opts.VolumeSizeMB > 0 {
		vm := opts.VolumeSizeMB
		body.VolumeSizeMB = &vm
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal compress body: %w", err)
	}

	endpoint := c.archiveURL(opts.Node, "compress")
	debugArchiveURL("compress", endpoint)
	respBody, err := c.do(ctx, http.MethodPost, endpoint, bytes.NewReader(raw), "application/json", password)
	if err != nil {
		return "", err
	}

	taskID, err := decodeCompressResponse(respBody)
	if err != nil {
		return "", fmt.Errorf("compress %s → %s: %w", strings.Join(opts.Sources, ","), opts.Destination, err)
	}
	return taskID, nil
}

// validateCompressOptions runs the client-side preflight on the
// option struct. Centralised so the cobra layer and any future
// programmatic caller (tests, an SDK) hit the same rules.
//
// Each refusal arm names the offending field AND the constraint
// it violates, so the error is self-describing without the cobra
// layer wrapping it.
func validateCompressOptions(opts CompressOptions, password string) error {
	if opts.Node == "" {
		return errors.New("compress: empty Node (cobra layer should resolve a default via /api/nodes/ or --node)")
	}
	if len(opts.Sources) == 0 {
		return errors.New("compress: at least one source path is required")
	}
	for _, s := range opts.Sources {
		if strings.TrimSpace(s) == "" {
			return errors.New("compress: source path is empty")
		}
	}
	if strings.TrimSpace(opts.Destination) == "" {
		return errors.New("compress: destination path is empty")
	}
	if err := ValidateFormat(opts.Format, "compress"); err != nil {
		return err
	}
	if err := ValidateLevel(opts.Level); err != nil {
		return err
	}
	if opts.VolumeSizeMB > 0 && !SupportsMultiVolume(opts.Format) {
		return fmt.Errorf(
			"compress: --volume-size-mb (%d) is only supported on multi-volume formats (zip, 7z); got format %q",
			opts.VolumeSizeMB, opts.Format)
	}
	if opts.VolumeSizeMB < 0 {
		return fmt.Errorf("compress: --volume-size-mb must be >= 0 (got %d)", opts.VolumeSizeMB)
	}
	if password != "" && !SupportsPassword(opts.Format) {
		return fmt.Errorf(
			"compress: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
			opts.Format)
	}
	if opts.Conflict != "" {
		// Defense in depth: ParseConflict at the cobra layer
		// has already validated this, but a programmatic caller
		// might construct the struct directly. Refuse unknown
		// values here so the contract holds end-to-end.
		ok := false
		for _, v := range validConflicts {
			if v == opts.Conflict {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("compress: invalid Conflict %q; valid values: %s",
				opts.Conflict, joinConflicts(validConflicts))
		}
	}
	return nil
}

// decodeCompressResponse extracts the task_id from a 2xx response
// body. The shared shape is `{code, msg, task_id}` — code != 0
// is treated as a server-side rejection (the malformed-path
// pattern the cp package's paste response also surfaces). An
// empty task_id is treated as protocol violation: a "queued"
// answer with no handle is useless and we refuse to silently
// drop it.
func decodeCompressResponse(body []byte) (string, error) {
	if len(body) == 0 {
		return "", errors.New("server returned empty body (expected {code, msg, task_id})")
	}
	var env compressResponseEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return "", fmt.Errorf("decode response: %w (body=%s)", err, truncateBody(body))
	}
	if env.Code != nil && *env.Code != 0 {
		msg := env.Msg
		if msg == "" {
			msg = fmt.Sprintf("server rejected the request (code %d)", *env.Code)
		}
		return "", errors.New(msg)
	}
	if env.TaskID == "" {
		return "", fmt.Errorf("server returned no task_id (body=%s)", truncateBody(body))
	}
	return env.TaskID, nil
}
