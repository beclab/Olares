package chart

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxExtractedBytes caps the total size an archive may expand to. Olares charts
// are a few kilobytes of templates; the cap only exists so a hostile .tgz cannot
// fill the disk of whatever runs lint or the preinstall image gate.
const MaxExtractedBytes int64 = 512 << 20

// MaxExtractedEntries caps how many members an archive may contain, for the same
// reason as MaxExtractedBytes.
const MaxExtractedEntries = 100_000

// ResolveDir turns a directory or a .tgz / .tar.gz package path into a chart
// directory that oac can consume. The returned cleanup must be called by the
// caller (via defer) and is always non-nil.
func ResolveDir(input string) (string, func(), error) {
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

	if err := ExtractTarGz(input, tmpDir); err != nil {
		cleanup()
		return "", noop, fmt.Errorf("extract %q: %w", input, err)
	}

	chartDir, err := LocateRoot(tmpDir)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	return chartDir, cleanup, nil
}

// ExtractTarGz unpacks a gzipped tar archive at src into dst, refusing any entry
// whose final destination escapes dst (zip-slip guard) and silently skipping
// non-regular non-directory entries (symlinks, devices, ...).
func ExtractTarGz(src, dst string) error {
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

	var (
		entries   int
		remaining = MaxExtractedBytes
	)
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
		entries++
		if entries > MaxExtractedEntries {
			return fmt.Errorf("archive holds more than %d entries", MaxExtractedEntries)
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
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			written, err := io.Copy(out, io.LimitReader(tr, remaining+1))
			out.Close()
			if err != nil {
				return err
			}
			if written > remaining {
				return fmt.Errorf("archive expands beyond %d bytes", MaxExtractedBytes)
			}
			remaining -= written
		default:
			// Skip symlinks / hardlinks / devices etc — Olares charts have no
			// legitimate use for them and accepting them is a foot-gun.
		}
	}
}

// LocateRoot picks the directory that should be passed to oac. Helm packaging
// puts every chart file under a single top-level directory named after the
// chart, so the common case is "exactly one subdirectory containing
// Chart.yaml". Chart.yaml directly at the extraction root is also accepted, for
// hand-rolled tarballs.
func LocateRoot(dir string) (string, error) {
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
