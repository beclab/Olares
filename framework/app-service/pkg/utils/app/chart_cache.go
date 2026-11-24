package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

// Chart cache layout (B):
//
//	./charts/<marketSource>/<owner>/<rawAppName>/<version>-<sha12>/<chart root>
//
// Every download is unpacked into a fresh directory named after the package
// digest and is never rewritten in place, so concurrent installs, upgrades and
// provider lookups of the same app can no longer overwrite each other's chart.
// The returned chart root travels as ApplicationConfig.ChartsName. Directories
// in the legacy flat layout (./charts/<rawAppName>) written by earlier
// releases are left untouched and keep working.

const (
	// maxChartPackageBytes caps a downloaded package; rendered charts are
	// kilobytes to a few megabytes.
	maxChartPackageBytes = 64 << 20
	// maxChartExpandedBytes caps what one package may expand to on disk.
	maxChartExpandedBytes = 256 << 20
	// chartVersionsKept and chartPruneAge bound the cache per
	// <source>/<owner>/<rawAppName>: a version directory is removed only when
	// it is outside the newest chartVersionsKept AND unused for chartPruneAge.
	// ApplicationManager.Spec.Config.ChartsName keeps pointing at the chart
	// an app was installed or upgraded from, and later operations (e.g. a
	// settings-triggered upgrade) read it again, so pruning is deliberately
	// conservative; readers that find their directory gone re-download it
	// (see prepare_upgrade.go).
	chartVersionsKept = 5
	chartPruneAge     = 30 * 24 * time.Hour
	// chartCompleteMarker is written last inside a version directory before
	// it is renamed into place; a directory without it is never reused.
	chartCompleteMarker  = ".complete"
	chartTmpDir          = ".tmp"
	chartDownloadTimeout = 60 * time.Second
)

// ErrChartDigestMismatch reports that the downloaded package does not match
// the digest market sent with the request.
var ErrChartDigestMismatch = errors.New("chart package digest mismatch")

// errChartPackageNotFound reports a 404 for the package URL.
var errChartPackageNotFound = errors.New("chart package not found")

var unsafePathChars = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// cacheSegment turns an identifier into one safe path segment.
func cacheSegment(s string) string {
	s = unsafePathChars.ReplaceAllString(s, "_")
	if s == "" || s == "." || s == ".." {
		return "_"
	}
	return s
}

// chartPackageRef identifies the package to fetch.
type chartPackageRef struct {
	url    string // absolute URL of the .tgz
	digest string // expected "sha256:<hex>"; empty when unknown (legacy index path)
}

// resolveChartURL resolves a chart URL sent by market against the repo URL and
// refuses anything that leaves the repo's scheme and host.
func resolveChartURL(repoURL, chartURL string) (string, error) {
	base, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("parse repo url: %w", err)
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	ref, err := url.Parse(chartURL)
	if err != nil {
		return "", fmt.Errorf("parse chart url: %w", err)
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != base.Scheme || resolved.Host != base.Host {
		return "", fmt.Errorf("chart url %q is outside repo %q", chartURL, repoURL)
	}
	return resolved.String(), nil
}

// chartRefMatches reports whether chartURL names the package of rawAppName at
// version (".../<raw>-<version>.tgz"), the only shape market sends.
func chartRefMatches(chartURL, rawAppName, version string) bool {
	u, err := url.Parse(chartURL)
	if err != nil {
		return false
	}
	return path.Base(u.Path) == rawAppName+"-"+version+".tgz"
}

func parseSHA256Digest(digest string) ([]byte, error) {
	hexPart, ok := strings.CutPrefix(digest, "sha256:")
	if !ok {
		return nil, fmt.Errorf("unsupported chart digest %q", digest)
	}
	sum, err := hex.DecodeString(hexPart)
	if err != nil || len(sum) != sha256.Size {
		return nil, fmt.Errorf("malformed chart digest %q", digest)
	}
	return sum, nil
}

// fetchChartPackage downloads the package bytes with the market headers.
func fetchChartPackage(ctx context.Context, options *ConfigOptions, packageURL string) ([]byte, error) {
	u, err := url.Parse(packageURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "file" {
		return readLimited(u.Path)
	}
	marketUser := options.Owner
	if options.ChartOwner != "" {
		marketUser = options.ChartOwner
	}
	resp, err := resty.New().SetTimeout(chartDownloadTimeout).
		SetAuthToken(options.Token).
		SetHeader(constants.MarketUser, marketUser).
		SetHeader(constants.MarketSource, options.MarketSource).
		R().SetContext(ctx).SetDoNotParseResponse(true).Get(packageURL)
	if err != nil {
		return nil, err
	}
	body := resp.RawBody()
	defer body.Close()
	if resp.StatusCode() == 404 {
		return nil, fmt.Errorf("download chart %s: %w", packageURL, errChartPackageNotFound)
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("download chart %s: unexpected status code %d", packageURL, resp.StatusCode())
	}
	data, err := io.ReadAll(io.LimitReader(body, maxChartPackageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("download chart %s: %w", packageURL, err)
	}
	if len(data) > maxChartPackageBytes {
		return nil, fmt.Errorf("download chart %s: package exceeds %d bytes", packageURL, maxChartPackageBytes)
	}
	return data, nil
}

func readLimited(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxChartPackageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxChartPackageBytes {
		return nil, fmt.Errorf("chart package %s exceeds %d bytes", p, maxChartPackageBytes)
	}
	return data, nil
}

// downloadChartToCache fetches ref, verifies its digest when one is given, and
// unpacks it into an immutable cache directory. It returns the chart root.
func downloadChartToCache(ctx context.Context, options *ConfigOptions, version string, ref chartPackageRef) (string, error) {
	data, err := fetchChartPackage(ctx, options, ref.url)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	if ref.digest != "" {
		want, err := parseSHA256Digest(ref.digest)
		if err != nil {
			return "", err
		}
		if !bytes.Equal(want, sum[:]) {
			return "", fmt.Errorf("%w: %s want %s got sha256:%x", ErrChartDigestMismatch, ref.url, ref.digest, sum)
		}
	}
	return storeChartPackage(appcfg.ChartsPath, options, version, hex.EncodeToString(sum[:]), data)
}

// storeChartPackage unpacks data under root and returns the chart root path.
// A complete directory with the same digest is reused as is.
func storeChartPackage(root string, options *ConfigOptions, version, sumHex string, data []byte) (string, error) {
	appDir := filepath.Join(root,
		cacheSegment(options.MarketSource),
		cacheSegment(options.Owner),
		cacheSegment(options.RawAppName))
	versionDir := filepath.Join(appDir, cacheSegment(version)+"-"+sumHex[:12])

	if _, err := os.Stat(filepath.Join(versionDir, chartCompleteMarker)); err == nil {
		// Reused: refresh its age so pruning treats it as recently used.
		now := time.Now()
		_ = os.Chtimes(filepath.Join(versionDir, chartCompleteMarker), now, now)
		return chartRootIn(versionDir, options.RawAppName)
	}

	tmpParent := filepath.Join(root, chartTmpDir)
	if err := os.MkdirAll(tmpParent, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(tmpParent, "chart-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp) // no-op once renamed

	if err := extractChartArchive(data, tmp); err != nil {
		return "", fmt.Errorf("unpack chart %s: %w", options.RawAppName, err)
	}
	if _, err := chartRootIn(tmp, options.RawAppName); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(tmp, chartCompleteMarker), []byte(sumHex+"\n"), 0o644); err != nil {
		return "", err
	}
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, versionDir); err != nil {
		// Another download of identical bytes won the race; its directory
		// is complete because the marker is written before the rename.
		if _, statErr := os.Stat(filepath.Join(versionDir, chartCompleteMarker)); statErr != nil {
			// An incomplete leftover (crash before rename can't produce
			// one, but be safe): replace it.
			if rmErr := os.RemoveAll(versionDir); rmErr != nil {
				return "", fmt.Errorf("replace incomplete chart dir: %w", rmErr)
			}
			if err := os.Rename(tmp, versionDir); err != nil {
				return "", err
			}
		}
	}
	pruneChartVersions(appDir, versionDir, chartVersionsKept, chartPruneAge, time.Now())
	return chartRootIn(versionDir, options.RawAppName)
}

// chartRootIn returns the chart root inside an unpacked package: the directory
// named after the raw app (what earlier releases assumed), else the archive's
// single top-level directory.
func chartRootIn(dir, rawAppName string) (string, error) {
	if rawAppName != "" {
		p := filepath.Join(dir, rawAppName)
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p, nil
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) != 1 {
		return "", fmt.Errorf("chart package for %q has %d top-level directories, want 1", rawAppName, len(dirs))
	}
	return filepath.Join(dir, dirs[0]), nil
}

// extractChartArchive unpacks a gzipped tar into dst. Only directories and
// regular files are written; absolute paths, ".." segments and links are
// refused, and the expanded size is capped.
func extractChartArchive(data []byte, dst string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var total int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := path.Clean(strings.ReplaceAll(hdr.Name, "\\", "/"))
		if name == "." || path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") {
			return fmt.Errorf("unsafe path %q in chart archive", hdr.Name)
		}
		target := filepath.Join(dst, filepath.FromSlash(name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			n, copyErr := io.Copy(f, io.LimitReader(tr, maxChartExpandedBytes-total+1))
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			total += n
			if total > maxChartExpandedBytes {
				return fmt.Errorf("chart archive expands beyond %d bytes", maxChartExpandedBytes)
			}
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			// metadata only
		default:
			return fmt.Errorf("unsupported entry %q (type %c) in chart archive", hdr.Name, hdr.Typeflag)
		}
	}
}

// pruneChartVersions removes complete version directories of one app that are
// outside the newest `keep` (current always counts as kept) and whose marker
// is older than maxAge. Failures are logged, never returned: pruning is
// housekeeping.
func pruneChartVersions(appDir, current string, keep int, maxAge time.Duration, now time.Time) {
	entries, err := os.ReadDir(appDir)
	if err != nil {
		return
	}
	type versionDir struct {
		path string
		mod  time.Time
	}
	var dirs []versionDir
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(appDir, e.Name())
		if p == current {
			continue
		}
		fi, err := os.Stat(filepath.Join(p, chartCompleteMarker))
		if err != nil {
			continue
		}
		dirs = append(dirs, versionDir{p, fi.ModTime()})
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].mod.After(dirs[j].mod) })
	for i, d := range dirs {
		if i < keep-1 || now.Sub(d.mod) < maxAge {
			continue
		}
		if err := os.RemoveAll(d.path); err != nil {
			klog.Warningf("prune chart cache %s: %v", d.path, err)
		}
	}
}
