// Package rag implements `olares-cli settings rag`, the authenticated CLI
// counterpart of Settings -> RAG. It calls user-service only; user-service is
// the trust boundary that injects the current user before forwarding Search3.
package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

type doer interface {
	DoJSON(context.Context, string, string, interface{}, interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    doer
}

type envelope struct {
	Code    int             `json:"code"`
	Message json.RawMessage `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings rag not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedClient{profile: rp, doer: whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)}, nil
}

func call(ctx context.Context, d doer, method, path string, body interface{}) (envelope, error) {
	return callAllowCodes(ctx, d, method, path, body)
}

// callAllowCodes keeps transport handling in one place while allowing an
// endpoint-specific empty-state code. Search3 returns NotFound (-2) with an
// empty array when a user has never configured RAG directories; for `dirs
// list` that is a valid empty result, not an operational failure.
func callAllowCodes(ctx context.Context, d doer, method, path string, body interface{}, allowed ...int) (envelope, error) {
	var response envelope
	if err := d.DoJSON(ctx, method, path, body, &response); err != nil {
		return response, err
	}
	accepted := response.Code == 0
	for _, code := range allowed {
		accepted = accepted || response.Code == code
	}
	if !accepted {
		message := envelopeMessage(response.Message)
		if message == "" {
			message = "request was not accepted"
		}
		return response, fmt.Errorf("RAG API code %d: %s", response.Code, message)
	}
	return response, nil
}

func envelopeMessage(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var message string
	if json.Unmarshal(raw, &message) == nil {
		return message
	}
	return strings.TrimSpace(string(raw))
}

func printData(data json.RawMessage, output string) error {
	if len(data) == 0 {
		data = json.RawMessage("null")
	}
	if output != "table" && output != "json" {
		return fmt.Errorf("unsupported --output %q (allowed: table, json)", output)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return fmt.Errorf("format RAG response: %w", err)
	}
	_, err := fmt.Fprintln(os.Stdout, buf.String())
	return err
}

func validateRoot(root string) (string, error) {
	trimmed := strings.TrimSpace(strings.TrimLeft(root, "/"))
	if trimmed == "" || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, ":") || strings.Contains(trimmed, "//") {
		return "", fmt.Errorf("invalid directory %q: use a frontend path such as drive/Home/Documents/", root)
	}
	for _, part := range strings.Split(trimmed, "/") {
		if part == "." || part == ".." {
			return "", fmt.Errorf("invalid directory %q: dot path segments are not allowed", root)
		}
	}
	return strings.TrimRight(trimmed, "/") + "/", nil
}
