package market

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// chartDownloadResult is the `-o json` shape of `market download`. It is not
// OperationResult: what a caller wants back here is where the file landed and
// how big it is, and neither has a home in the lifecycle result type.
type chartDownloadResult struct {
	App       string `json:"app"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path"`
	Bytes     int64  `json:"bytes"`
	User      string `json:"user,omitempty"`
}

func NewCmdMarketDownload(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	overwrite := false
	cmd := &cobra.Command{
		Use:     "download {app-name} [local-path]",
		Aliases: []string{"pull"},
		Short:   "Download an app's chart package (.tgz) from a market source",
		Long: `Download the chart package the market holds for an app.

This is the read side of 'market upload': it fetches the exact .tgz the
market serves to the installer, so a chart that only exists on the
Olares side — the local working copy was deleted, or the upload happened
from another machine — can be recovered and edited, re-packaged and
re-uploaded.

Source defaults to the SPA's "Local Sources → Upload" bucket ('upload'),
the same bucket 'upload' writes to and 'delete' removes from. Pass
-s/--source to read a chart out of another source (e.g. market.olares)
when that source's chart has been cached on this Olares.

If --version is omitted the market serves the version it currently holds
for the app, which is what you want when recovering a chart: it does not
depend on the app being installed, or on the install having succeeded.

Local destination semantics mirror 'olares-cli files download':

  omitted        -> ./<chart>-<version>.tgz (the name the server suggests)
  existing dir   -> that directory, using the server-suggested name
  any other path -> the full local target path

An existing local file is not overwritten unless --overwrite is passed;
the download writes to <path>.tmp and renames on success, so a failure
part-way through leaves any previous chart intact.

Examples:
  olares-cli market download myapp                     # ./myapp-1.0.0.tgz
  olares-cli market download myapp ./charts/           # into a directory
  olares-cli market download myapp ./myapp.tgz         # exact filename
  olares-cli market download myapp --version 1.0.0     # a specific version
  olares-cli market download myapp -s market.olares    # another source
  olares-cli market download myapp -o json             # structured result`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			localArg := ""
			if len(args) == 2 {
				localArg = args[1]
			}
			return runDownload(cmd.Context(), opts, args[0], localArg, overwrite)
		},
	}
	opts.addSourceFlag(cmd, fmt.Sprintf("market source id to read the chart from (default %q)", chartUploadSource))
	opts.addVersionFlag(cmd)
	opts.addOutputFlags(cmd)
	cmd.Flags().BoolVar(&overwrite, "overwrite", false,
		"replace an existing local file (writes to <path>.tmp + rename)")
	return cmd
}

func runDownload(ctx context.Context, opts *MarketOptions, appName, localArg string, overwrite bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("download", appName, err)
	}

	source := strings.TrimSpace(opts.Source)
	if source == "" {
		source = chartUploadSource
		opts.info("Using source: %s", source)
	}

	version := strings.TrimSpace(opts.Version)
	if version != "" {
		if err := validateVersion(version); err != nil {
			return opts.failOp("download", appName, err)
		}
	}

	result, err := fetchChartToDisk(ctx, mc, appName, source, version, localArg, overwrite)
	if err != nil {
		return opts.failOp("download", appName, err)
	}
	result.User = mc.olaresID

	if opts.Quiet {
		return nil
	}
	if opts.isJSON() {
		return opts.printJSON(result)
	}
	fmt.Fprintf(os.Stdout, "download '%s': wrote %d bytes to %s\n", result.App, result.Bytes, result.Path)
	fmt.Fprintf(os.Stdout, "  source: %s\n", result.Source)
	if result.Version != "" {
		fmt.Fprintf(os.Stdout, "  version: %s\n", result.Version)
	}
	return nil
}

// fetchChartToDisk streams the chart into its local destination and reports
// what landed there. Split out of runDownload so the wire + filesystem
// behaviour is testable without a resolved profile.
func fetchChartToDisk(ctx context.Context, mc *MarketClient, appName, source, version, localArg string, overwrite bool) (chartDownloadResult, error) {
	result := chartDownloadResult{
		App:       appName,
		Operation: "download",
		Status:    "success",
		Source:    source,
		Version:   version,
	}

	pkg, err := mc.DownloadChart(ctx, appName, source, version)
	if err != nil {
		return result, err
	}
	defer pkg.Body.Close()

	if result.Version == "" {
		result.Version = versionFromChartFilename(pkg.Filename, appName)
	}
	dst, err := resolveChartDestination(localArg, pkg.Filename, appName, result.Version)
	if err != nil {
		return result, err
	}
	written, err := writeChartFile(dst, pkg.Body, overwrite)
	if err != nil {
		return result, err
	}
	result.Path = dst
	result.Bytes = written
	return result, nil
}

// resolveChartDestination applies the same local-path rules as
// `files download` (see resolveLocalFile there): an omitted or
// directory-shaped argument means "use the server's filename in that
// directory", anything else is the full target path.
func resolveChartDestination(localArg, serverFilename, appName, version string) (string, error) {
	base := strings.TrimSpace(serverFilename)
	if base == "" || base != filepath.Base(base) || base == "." || base == ".." {
		// No usable suggestion, or one carrying path separators — a
		// server-chosen name must never be able to steer the write out of
		// the directory the user named.
		base = defaultChartFilename(appName, version)
	}
	if localArg == "" {
		return base, nil
	}
	st, err := os.Stat(localArg)
	switch {
	case err == nil && st.IsDir():
		return filepath.Join(localArg, base), nil
	case err == nil:
		return localArg, nil
	case errors.Is(err, os.ErrNotExist):
		// Trailing slash means "a directory, even though it does not exist
		// yet" — the `cp` / `rsync` convention.
		if strings.HasSuffix(localArg, "/") || strings.HasSuffix(localArg, string(os.PathSeparator)) {
			return filepath.Join(localArg, base), nil
		}
		return localArg, nil
	default:
		return "", fmt.Errorf("stat %s: %w", localArg, err)
	}
}

func defaultChartFilename(appName, version string) string {
	if version == "" {
		return appName + ".tgz"
	}
	return fmt.Sprintf("%s-%s.tgz", appName, version)
}

// versionFromChartFilename recovers the version from the server's
// `<chart>-<version>.tgz` suggestion. Chart names contain hyphens, so this
// only trusts a filename that starts with the app name; anything else leaves
// the version unreported rather than guessed.
func versionFromChartFilename(filename, appName string) string {
	base := strings.TrimSuffix(strings.TrimSpace(filename), ".tgz")
	if base == "" || !strings.HasPrefix(base, appName+"-") {
		return ""
	}
	return strings.TrimPrefix(base, appName+"-")
}

// writeChartFile streams the body through <dst>.tmp and renames it into
// place, so an interrupted transfer cannot leave a truncated chart where a
// good one used to be.
func writeChartFile(dst string, body io.Reader, overwrite bool) (int64, error) {
	if _, err := os.Stat(dst); err == nil {
		if !overwrite {
			return 0, fmt.Errorf("%s already exists (pass --overwrite to replace it)", dst)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("stat %s: %w", dst, err)
	}

	if parent := filepath.Dir(dst); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return 0, fmt.Errorf("mkdir parent of %s: %w", dst, err)
		}
	}

	tmp := dst + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", tmp, err)
	}
	written, copyErr := io.Copy(f, body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("write %s: %w", tmp, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("close %s: %w", tmp, closeErr)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("rename %s -> %s: %w", tmp, dst, err)
	}
	return written, nil
}

// parseContentDispositionFilename pulls the basename out of a
// `Content-Disposition: attachment; filename="<basename>"` header.
//
// Mirrors the helper of the same name in internal/files/archive: a
// best-effort substring extractor rather than mime.ParseMediaType, which
// rejects the partially-escaped filenames these backends emit. The two are
// separate copies because a string helper is not worth a dependency from the
// market verbs onto the files internals.
func parseContentDispositionFilename(cd string) string {
	if cd == "" {
		return ""
	}
	const key = "filename="
	idx := strings.Index(strings.ToLower(cd), key)
	if idx < 0 {
		return ""
	}
	val := strings.TrimSpace(cd[idx+len(key):])
	if strings.HasPrefix(val, "\"") {
		if end := strings.Index(val[1:], "\""); end >= 0 {
			return val[1 : 1+end]
		}
		return val[1:]
	}
	if i := strings.Index(val, ";"); i >= 0 {
		val = val[:i]
	}
	return strings.TrimSpace(val)
}
