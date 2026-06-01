// Package containerdimages is the shared wire model + fetch path for
// the containerd image list exposed by user-service at
// /api/containerd/images (a BFL-enveloped array). Both
// `olares-cli settings advanced images list` and
// `olares-cli doctor images` decode the same shape, so the type
// and the GET live here instead of being duplicated per command.
package containerdimages

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Image is one entry from /api/containerd/images.
//
// RepoDigests carries the registry-canonical digest references
// (repo@sha256:...) when the daemon reports them; it lets digest-pinned
// workload references match a local image even when no tag does.
type Image struct {
	ID          string   `json:"id"`
	Size        int64    `json:"size"`
	RepoTags    []string `json:"repo_tags"`
	RepoDigests []string `json:"repo_digests,omitempty"`
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// List fetches containerd images, optionally scoped to a single
// registry. The request rides the desktop origin: both the desktop and
// settings nginx fronts forward /api/* to the same user-service
// upstream (see pkg/olares/id.go DesktopURL).
func List(ctx context.Context, f *cmdutil.Factory, registry string) ([]Image, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return nil, fmt.Errorf("internal error: containerdimages.List not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	path := "/api/containerd/images"
	if registry != "" {
		path += "?registry=" + url.QueryEscape(registry)
	}
	client := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)
	var env envelope
	if err := client.DoJSON(ctx, http.MethodGet, path, nil, &env); err != nil {
		return nil, err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return nil, fmt.Errorf("GET %s: upstream returned code %d", path, env.Code)
		}
		return nil, fmt.Errorf("GET %s: upstream returned code %d: %s", path, env.Code, msg)
	}
	var images []Image
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &images); err != nil {
			return nil, fmt.Errorf("GET %s: decode data: %w", path, err)
		}
	}
	return images, nil
}

// HumanBytes renders byte counts as KiB/MiB/GiB using base-2.
func HumanBytes(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffix := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), suffix)
}

// ShortID trims the sha256: prefix and keeps the first 12 chars, like
// the SPA's image table.
func ShortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
