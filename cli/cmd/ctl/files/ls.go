package files

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type lsOptions struct {
	asJSON bool
}

// NewLsCommand: `olares-cli files ls <frontendPath> [--json]`
//
// Calls GET <FilesURL>/api/resources/<fileType>/<extend><subPath> on the
// per-user files-backend (proxied via files.<terminusName>) and renders the
// result. The access token is injected by Factory's HTTP client as the
// `X-Authorization` header — see pkg/cmdutil/factory.go for why that header
// (not the standard Authorization: Bearer) is the right one for Olares.
//
// Errors:
//   - bad / missing path is rejected client-side via ParseFrontendPath
//   - 401/403 from the backend is reported with the same "run profile login"
//     CTA that DefaultProvider uses, so the message is consistent across
//     "no token" / "expired token" / "server-rejected token"
//   - other non-2xx responses surface the backend's error/message JSON field
//     verbatim, which is usually enough to debug (unknown node, missing repo,
//     permission denied, ...)
func NewLsCommand(f *cmdutil.Factory) *cobra.Command {
	o := &lsOptions{}
	cmd := &cobra.Command{
		Use:   "ls <frontendPath>",
		Short: "list a directory on the per-user files-backend",
		Long: `List a directory on the per-user files-backend.

The path is the full 3-segment front-end path used by the backend
(<fileType>/<extend>[/<subPath>]); see ` + "`olares-cli files --help`" + ` for
the schema.

Examples:

    olares-cli files ls drive/Home/
    olares-cli files ls drive/Home/Documents
    olares-cli files ls drive/Data/
    olares-cli files ls cache/<node>/
    olares-cli files ls sync/<repo_id>/
    olares-cli files ls awss3/<account>/<bucket>
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLs(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().BoolVar(&o.asJSON, "json", false, "print the raw JSON response (pretty-printed) instead of a table")
	return cmd
}

// listingItem is a deliberately-narrow projection of files-backend's
// FileInfo (files/pkg/files/file.go). We don't import the backend struct
// directly to avoid pulling in afero / klog / the rest of the per-user
// service into the CLI binary; we only decode what we render in the
// table view (MODE / SIZE / TYPE / MODIFIED / NAME) plus Path, which is
// handy for diagnostic error messages and for the future cat/cp/rm verbs.
//
// `Mode` is the raw integer value of Go's os.FileMode (ModeDir | perms |
// ...) — the backend marshals it that way, see files/pkg/files/file.go.
// `Type` is the backend's semantic class (one of "" / blob / video /
// audio / image / pdf / text / textImmutable / invalid_link); empty for
// directories. We pass it through verbatim and let the user see the same
// label the web app would.
type listingItem struct {
	Name      string    `json:"name"`
	IsDir     bool      `json:"isDir"`
	IsSymlink bool      `json:"isSymlink"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	Mode      uint32    `json:"mode"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
}

// listingResponse decodes both the parent-directory envelope (used to print
// a one-line header before the table) and the items it contains. NumDirs /
// NumFiles come from the backend; we use them verbatim when present and
// fall back to counting `Items` if the backend reports zeros (defensive —
// older response shapes may not populate them for every fileType).
type listingResponse struct {
	Name      string        `json:"name"`
	Path      string        `json:"path"`
	Modified  time.Time     `json:"modified"`
	Mode      uint32        `json:"mode"`
	IsSymlink bool          `json:"isSymlink"`
	NumDirs   int           `json:"numDirs"`
	NumFiles  int           `json:"numFiles"`
	Items     []listingItem `json:"items"`
}

func runLs(ctx context.Context, f *cmdutil.Factory, out io.Writer, rawPath string, o *lsOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}

	fp, err := ParseFrontendPath(rawPath)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	client, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}

	// URLPath uses upload.EncodeURL (same as download/cat/rm/upload) so
	// filenames with '#', '?', '+', spaces, '!*'()', etc. survive
	// the trip to the backend. ParseFrontendPath already guarantees that
	// listing the extend root ("drive/Home" or "drive/Home/") yields a
	// SubPath of "/", so URLPath() naturally ends with '/' there — which is
	// what FileParam.convert() in files-backend requires
	// (it rejects len(strings.Split(u, "/")) < 3).
	endpoint := rp.FilesURL + "/api/resources/" + fp.URLPath()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return formatHTTPError(resp.StatusCode, body, rp.OlaresID, endpoint)
	}

	if o.asJSON {
		return prettyPrintJSON(out, body)
	}

	var listing listingResponse
	if err := json.Unmarshal(body, &listing); err != nil {
		return fmt.Errorf("decode response: %w (body=%s)", err, truncate(string(body), 200))
	}
	return renderListing(out, fp, listing)
}

// formatHTTPError turns a non-2xx response into a user-facing error. 401/403
// is special-cased to match DefaultProvider's CTA so the user sees the same
// hint whether the local check or the remote check is what failed.
func formatHTTPError(status int, body []byte, olaresID, url string) error {
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
			status, olaresID)
	}
	// Backend returns errors as either {"error": "..."} or {"code":1,"message":"..."}.
	// Try to surface either; fall back to the raw body.
	var generic struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	if err := json.Unmarshal(body, &generic); err == nil {
		switch {
		case generic.Error != "":
			return fmt.Errorf("GET %s: HTTP %d: %s", url, status, generic.Error)
		case generic.Message != "":
			return fmt.Errorf("GET %s: HTTP %d (code=%d): %s", url, status, generic.Code, generic.Message)
		}
	}
	return fmt.Errorf("GET %s: HTTP %d: %s", url, status, truncate(string(body), 500))
}

// renderListing prints (a) a one-line header summarising the directory the
// user just listed, and (b) a 5-column table of its contents
// (MODE / SIZE / TYPE / MODIFIED / NAME). Directories sort first, then
// files, both case-insensitive alphabetical. Directory names get a
// trailing '/' so the distinction is also visible per row.
//
// Empty directories print the header followed by "(empty)" — the header is
// always present so the user sees the directory's own modified-time and
// dir/file counts even when there's nothing inside.
func renderListing(w io.Writer, fp FrontendPath, listing listingResponse) error {
	writeListingHeader(w, fp, listing)

	if len(listing.Items) == 0 {
		_, err := fmt.Fprintln(w, "(empty)")
		return err
	}

	items := append([]listingItem(nil), listing.Items...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir // dirs first
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "MODE\tSIZE\tTYPE\tMODIFIED\tNAME")
	for _, it := range items {
		modified := "-"
		if !it.Modified.IsZero() {
			modified = it.Modified.Local().Format("2006-01-02 15:04")
		}
		name := it.Name
		if it.IsDir {
			name += "/"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			formatMode(it.Mode, it.IsDir, it.IsSymlink),
			formatSize(it.Size, it.IsDir),
			formatType(it.Type, it.IsDir),
			modified,
			name,
		)
	}
	return tw.Flush()
}

// writeListingHeader prints a single banner line of the form
//
//	drive/Home/Code  (1 dir, 3 files, modified 2026-04-17 19:31)
//
// Counts come from the envelope; we fall back to counting items when the
// backend reports zeros but the listing clearly isn't empty (defensive).
// The "modified" suffix is omitted when the envelope didn't carry one.
func writeListingHeader(w io.Writer, fp FrontendPath, listing listingResponse) {
	dirs, files := listing.NumDirs, listing.NumFiles
	if dirs == 0 && files == 0 && len(listing.Items) > 0 {
		for _, it := range listing.Items {
			if it.IsDir {
				dirs++
			} else {
				files++
			}
		}
	}

	parts := []string{
		pluralize(dirs, "dir", "dirs"),
		pluralize(files, "file", "files"),
	}
	if !listing.Modified.IsZero() {
		parts = append(parts, "modified "+listing.Modified.Local().Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(w, "%s  (%s)\n", fp.String(), strings.Join(parts, ", "))
}

// pluralize returns "<n> <singular|plural>". Tiny helper, but it makes the
// header read naturally for the common 0/1/many cases ("0 dirs, 1 file").
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

// formatSize renders bytes in a compact human-friendly form (1.2K, 3.4M).
// Directories report "-" because their backend Size is meaningless without
// a recursive walk and would confuse users.
func formatSize(n int64, isDir bool) string {
	if isDir {
		return "-"
	}
	return formatBytes(n)
}

// formatBytes renders a byte count for CLI progress lines (ls rows use
// formatSize; upload/download share this helper).
func formatBytes(n int64) string {
	const unit = 1024
	if n < 0 {
		return fmt.Sprintf("%dB", n)
	}
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for n2 := n / unit; n2 >= unit; n2 /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// formatMode renders the per-row mode column. The backend ships os.FileMode
// as a raw integer (ModeDir | perms | ...); when it's nonzero we delegate
// to os.FileMode.String(), which gives us proper "drwxr-xr-x" / "Lrwxr-xr-x"
// / "-rw-r--r--" forms — a strict superset of the old "-"/"d" indicator.
//
// When `mode` is zero (older response shapes / partial fixtures) we still
// surface dir/symlink-ness from the dedicated bool fields so the column
// remains informative.
func formatMode(mode uint32, isDir, isSymlink bool) string {
	if mode != 0 {
		return os.FileMode(mode).String()
	}
	switch {
	case isSymlink:
		return "L---------"
	case isDir:
		return "d---------"
	default:
		return "----------"
	}
}

// formatType returns what to display in the TYPE column. The backend's empty
// string is rendered as "-" so the column stays visually aligned for
// directories and uncategorised entries.
func formatType(t string, isDir bool) string {
	if isDir || t == "" {
		return "-"
	}
	return t
}

func prettyPrintJSON(w io.Writer, body []byte) error {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		_, werr := w.Write(body)
		return werr
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
