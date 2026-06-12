// extract.go: wire side of POST /api/archive/<node>/extract.
//
// JSON body shape (mirrors the spec verbatim — keep in sync if
// the backend ever renames a field):
//
//	{
//	  "source":           "/drive/Home/out.zip",
//	  "destination":      "/drive/Home/unpacked",
//	  "format":           "zip",
//	  "preserveSymlinks": false,
//	  "conflict":         "rename"
//	}
//
// Same task-queue model as compress: the response only confirms
// "task is queued"; the actual extraction runs asynchronously
// and the caller polls /api/task/<node>/?task_id=... for
// completion via the WaitTask helper.
//
// Password handling is identical — sent through the
// `X-Archive-Password` header, only valid for zip / 7z.
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

// ExtractOptions captures the knobs the extract endpoint
// accepts. Construct via the cobra layer's option parser.
//
// Format is required — the backend uses it to pick the right
// decoder. The cobra layer's default heuristic (derive from
// Source basename via FormatFromExtension) keeps this from
// being a routine annoyance; the user only needs to type
// --format when the heuristic fails (e.g. an archive without a
// canonical extension).
type ExtractOptions struct {
	// Source is the wire path of the archive to extract.
	// Must exist on the server.
	Source string

	// Destination is the wire path of the directory to extract
	// INTO. Per the spec, the backend creates intermediate
	// directories as needed; the conflict policy decides what
	// happens when a child of the archive collides with an
	// existing entry at the destination.
	Destination string

	// Format is the validated archive format (one of
	// SupportedFormats). The cobra layer ensures this is set
	// before reaching here.
	Format string

	// PreserveSymlinks: when true, symlinks inside the archive
	// land as symlinks on disk; false dereferences them at
	// extraction time. False matches the spec's default.
	PreserveSymlinks bool

	// Conflict is the on-collision policy for entries that
	// would overwrite an existing path under Destination.
	// Empty → ConflictDefault (rename) before send.
	Conflict Conflict

	// Node is the {node} URL segment. Same resolution as
	// compress.
	Node string
}

// extractRequestBody is the JSON shape the backend's extract
// handler binds to. The shape is intentionally minimal — extract
// is "given this archive and this dir, undo the compress" and
// most knobs from compress (level, volumeSizeMB) don't apply.
type extractRequestBody struct {
	Source           string `json:"source"`
	Destination      string `json:"destination"`
	Format           string `json:"format"`
	PreserveSymlinks bool   `json:"preserveSymlinks"`
	Conflict         string `json:"conflict,omitempty"`
}

// extractResponseEnvelope mirrors the spec's success response:
//
//	{ "code": 0, "msg": "success", "task_id": "<id>" }
//
// Same shape as compress's response — we keep the type separate
// so a future schema divergence between the two endpoints
// surfaces as a compile-time field difference rather than a
// silent runtime mistake.
type extractResponseEnvelope struct {
	Code   *int   `json:"code,omitempty"`
	Msg    string `json:"msg,omitempty"`
	TaskID string `json:"task_id,omitempty"`
}

// Extract sends one POST /api/archive/<node>/extract for the
// supplied options and returns the resulting task_id. Same
// async semantics as Compress — by the time this returns the
// extraction has only been queued, not completed.
//
// Validation runs client-side first; HTTP errors surface as
// *HTTPError for the reformatter.
func (c *Client) Extract(ctx context.Context, opts ExtractOptions, password string) (string, error) {
	if err := validateExtractOptions(opts, password); err != nil {
		return "", err
	}

	body := extractRequestBody{
		Source:           opts.Source,
		Destination:      opts.Destination,
		Format:           strings.ToLower(opts.Format),
		PreserveSymlinks: opts.PreserveSymlinks,
	}
	if opts.Conflict != "" {
		body.Conflict = string(opts.Conflict)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal extract body: %w", err)
	}

	endpoint := c.archiveURL(opts.Node, "extract")
	debugArchiveURL("extract", endpoint)
	respBody, err := c.do(ctx, http.MethodPost, endpoint, bytes.NewReader(raw), "application/json", password)
	if err != nil {
		return "", err
	}

	taskID, err := decodeExtractResponse(respBody)
	if err != nil {
		return "", fmt.Errorf("extract %s → %s: %w", opts.Source, opts.Destination, err)
	}
	return taskID, nil
}

// validateExtractOptions mirrors validateCompressOptions's
// fast-fail spirit on the smaller extract surface.
func validateExtractOptions(opts ExtractOptions, password string) error {
	if opts.Node == "" {
		return errors.New("extract: empty Node (cobra layer should resolve a default via /api/nodes/ or --node)")
	}
	if strings.TrimSpace(opts.Source) == "" {
		return errors.New("extract: source archive path is empty")
	}
	if strings.TrimSpace(opts.Destination) == "" {
		return errors.New("extract: destination directory path is empty")
	}
	if err := ValidateFormat(opts.Format, "extract"); err != nil {
		return err
	}
	if password != "" && !SupportsPassword(opts.Format) {
		return fmt.Errorf(
			"extract: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
			opts.Format)
	}
	if opts.Conflict != "" {
		ok := false
		for _, v := range validConflicts {
			if v == opts.Conflict {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("extract: invalid Conflict %q; valid values: %s",
				opts.Conflict, joinConflicts(validConflicts))
		}
	}
	return nil
}

// decodeExtractResponse extracts the task_id from a 2xx response
// body — same shape and semantics as decodeCompressResponse but
// kept as its own function so a future schema split between the
// two endpoints lands on a compile-time field error instead of a
// silent runtime mismatch.
func decodeExtractResponse(body []byte) (string, error) {
	if len(body) == 0 {
		return "", errors.New("server returned empty body (expected {code, msg, task_id})")
	}
	var env extractResponseEnvelope
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
