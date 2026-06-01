package chart

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	oac "github.com/beclab/Olares/framework/oac"
	"github.com/spf13/cobra"
)

type lintOpts struct {
	Owner     string
	Admin     string
	AutoOwner bool

	SkipFolder      bool
	SkipManifest    bool
	SkipResource    bool
	SkipHostPath    bool
	SkipNamespace   bool
	SkipAppData     bool
	SkipSameVersion bool

	WithRBAC            bool
	WithSecurityContext bool
}

func NewCmdChartLint() *cobra.Command {
	o := &lintOpts{AutoOwner: true}
	cmd := &cobra.Command{
		Use:   "lint <path>",
		Short: "Validate a directory or .tgz/.tar.gz as a valid Olares chart",
		Long: `Validate that a chart directory or .tgz / .tar.gz package is a valid
Olares chart. The check runs the same pipeline the app store uses to
ingest a chart:

  - chart folder layout (Chart.yaml / values.yaml / templates / OlaresManifest.yaml)
  - OlaresManifest.yaml structural + cross-field validation
  - helm dry-run + workload-integrity, hostPath, namespace checks
  - container-level resource-limit checks
  - Chart.yaml <-> OlaresManifest.yaml version consistency

By default the chart is rendered under both owner==admin and owner!=admin
install scenarios; use --owner / --admin (with --auto-owner=false) to pin
a specific scenario instead.

Examples:
  olares-cli chart lint ./my-app
  olares-cli chart lint ./my-app-1.0.0.tgz
  olares-cli chart lint ./my-app --skip-resource --with-rbac
  olares-cli chart lint ./my-app --auto-owner=false --owner alice --admin root`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(args[0], o)
		},
	}
	fs := cmd.Flags()
	fs.StringVar(&o.Owner, "owner", "", "set .Values.bfl.username for the helm dry-run (overridden by --auto-owner)")
	fs.StringVar(&o.Admin, "admin", "", "set .Values.admin for the helm dry-run (overridden by --auto-owner)")
	fs.BoolVar(&o.AutoOwner, "auto-owner", true, "lint under both owner==admin and owner!=admin scenarios; ignores --owner/--admin when on")

	fs.BoolVar(&o.SkipFolder, "skip-folder", false, "skip folder-layout check")
	fs.BoolVar(&o.SkipManifest, "skip-manifest", false, "skip OlaresManifest.yaml structural validation")
	fs.BoolVar(&o.SkipResource, "skip-resource", false, "skip container-level resource-limit checks")
	fs.BoolVar(&o.SkipHostPath, "skip-host-path", false, "skip hostPath + rolling-update incompatibility check")
	fs.BoolVar(&o.SkipNamespace, "skip-namespace", false, "skip rendered-resource namespace check")
	fs.BoolVar(&o.SkipAppData, "skip-app-data", false, "skip .Values.userspace.appdata cross-check")
	fs.BoolVar(&o.SkipSameVersion, "skip-same-version", false, "skip Chart.yaml <-> OlaresManifest.yaml version match")

	fs.BoolVar(&o.WithRBAC, "with-rbac", false, "enable ServiceAccount RBAC forbidden-rules check (off by default)")
	fs.BoolVar(&o.WithSecurityContext, "with-security-context", false, "enable non-beclab privileged securityContext check (off by default)")

	return cmd
}

func runLint(input string, o *lintOpts) error {
	chartDir, cleanup, err := resolveChartDir(input)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := oac.Lint(chartDir, buildOACOptions(o)...); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s: OK\n", input)
	return nil
}

func buildOACOptions(o *lintOpts) []oac.Option {
	opts := []oac.Option{}
	if o.AutoOwner {
		opts = append(opts, oac.WithAutoOwnerScenarios())
	} else {
		if o.Owner != "" {
			opts = append(opts, oac.WithOwner(o.Owner))
		}
		if o.Admin != "" {
			opts = append(opts, oac.WithAdmin(o.Admin))
		}
	}
	if o.SkipFolder {
		opts = append(opts, oac.SkipFolderCheck())
	}
	if o.SkipManifest {
		opts = append(opts, oac.SkipManifestCheck())
	}
	if o.SkipResource {
		opts = append(opts, oac.SkipResourceCheck())
	}
	if o.SkipHostPath {
		opts = append(opts, oac.SkipHostPathCheck())
	}
	if o.SkipNamespace {
		opts = append(opts, oac.SkipResourceNamespaceCheck())
	}
	if o.SkipAppData {
		opts = append(opts, oac.SkipAppDataCheck())
	}
	if o.SkipSameVersion {
		opts = append(opts, oac.SkipSameVersionCheck())
	}
	if o.WithRBAC {
		opts = append(opts, oac.WithServiceAccountRulesCheck())
	}
	if o.WithSecurityContext {
		opts = append(opts, oac.WithSecurityContextCheck())
	}
	return opts
}

// resolveChartDir turns a directory or a .tgz / .tar.gz package path into a
// chart directory that oac.Lint can consume. The returned cleanup must be
// called by the caller (via defer) and is always non-nil.
func resolveChartDir(input string) (string, func(), error) {
	noop := func() {}
	info, err := os.Stat(input)
	if err != nil {
		return "", noop, fmt.Errorf("cannot access %q: %w", input, err)
	}
	if info.IsDir() {
		return input, noop, nil
	}
	lower := strings.ToLower(info.Name())
	if !strings.HasSuffix(lower, ".tgz") && !strings.HasSuffix(lower, ".tar.gz") {
		return "", noop, fmt.Errorf("unsupported file format: expected directory, .tgz, or .tar.gz")
	}

	tmpDir, err := os.MkdirTemp("", "olares-chart-*")
	if err != nil {
		return "", noop, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	if err := extractTarGz(input, tmpDir); err != nil {
		cleanup()
		return "", noop, fmt.Errorf("extract %q: %w", input, err)
	}

	chartDir, err := locateChartRoot(tmpDir)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	return chartDir, cleanup, nil
}

// extractTarGz unpacks a gzipped tar archive at src into dst, refusing any
// entry whose final destination escapes dst (zip-slip guard) and silently
// skipping non-regular non-directory entries (symlinks, devices, ...).
func extractTarGz(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr == nil || hdr.Name == "" {
			continue
		}

		// zip-slip: reject ".." traversal and absolute paths.
		clean := filepath.Clean(hdr.Name)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
			return fmt.Errorf("invalid tar entry %q: escapes archive root", hdr.Name)
		}
		target := filepath.Join(dstAbs, clean)
		if !strings.HasPrefix(target, dstAbs+string(filepath.Separator)) && target != dstAbs {
			return fmt.Errorf("invalid tar entry %q: escapes archive root", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		default:
			// Skip symlinks / hardlinks / devices etc — Olares charts have no
			// legitimate use for them and accepting them is a foot-gun.
		}
	}
}

// locateChartRoot picks the directory that should be passed to oac.Lint.
// Helm packaging puts every chart file under a single top-level directory
// named after the chart, so the common case is "exactly one subdirectory
// containing Chart.yaml". We also accept Chart.yaml directly at the
// extraction root for hand-rolled tarballs.
func locateChartRoot(dir string) (string, error) {
	if fileExists(filepath.Join(dir, "Chart.yaml")) {
		return dir, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var subDirs []string
	for _, e := range entries {
		if e.IsDir() {
			subDirs = append(subDirs, e.Name())
		}
	}
	if len(subDirs) == 1 {
		candidate := filepath.Join(dir, subDirs[0])
		if fileExists(filepath.Join(candidate, "Chart.yaml")) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("tarball does not contain a single chart root (no Chart.yaml found at the expected location)")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
