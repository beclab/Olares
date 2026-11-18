package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// CleanCache removes cached packages and image archives absent from the selected manifest.
func CleanCache(ctx context.Context, baseDir, cacheDir, release string, dryRun bool, out io.Writer) error {
	if _, err := semver.StrictNewVersion(release); err != nil {
		return fmt.Errorf("invalid Olares version %q: %w", release, err)
	}
	installation, err := os.OpenRoot(baseDir)
	if err != nil {
		return fmt.Errorf("open Olares base directory: %w", err)
	}
	defer installation.Close()

	keep, err := readCacheManifest(installation, "v"+release)
	if err != nil {
		return err
	}

	cachePath, err := filepath.Abs(cacheDir)
	if err != nil {
		return fmt.Errorf("resolve Olares cache directory: %w", err)
	}
	cachePath, err = filepath.EvalSymlinks(cachePath)
	if err != nil {
		return fmt.Errorf("resolve Olares cache directory: %w", err)
	}
	cache, err := os.OpenRoot(cachePath)
	if err != nil {
		return fmt.Errorf("open Olares cache directory: %w", err)
	}
	defer cache.Close()

	protected, err := protectedCachePaths(cachePath, keep)
	if err != nil {
		return err
	}
	candidates, err := findUnusedCacheFiles(ctx, cache.FS(), protected)
	if err != nil {
		return fmt.Errorf("scan download cache: %w", err)
	}

	action := "Removed"
	if dryRun {
		action = "Would remove"
	}
	var count int
	var size int64
	for _, name := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := regularCacheFile(cache, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect cached file %q: %w", name, err)
		}
		if info == nil {
			continue
		}
		if !dryRun {
			if err := cache.Remove(name); errors.Is(err, fs.ErrNotExist) {
				continue
			} else if err != nil {
				return fmt.Errorf("remove cached file %q: %w", name, err)
			}
		}
		count++
		size += info.Size()
		if _, err := fmt.Fprintf(out, "%s %s\n", action, filepath.Join(cacheDir, name)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "%s %d cached files (%s).\n", action, count, utils.FormatBytes(size))
	return err
}

func protectedCachePaths(cachePath string, keep []string) (map[string]struct{}, error) {
	protected := make(map[string]struct{}, len(keep))
	for _, name := range keep {
		// Keep case variants on case-insensitive filesystems.
		protected[strings.ToLower(name)] = struct{}{}

		// A manifest entry can itself be a symlink to a regular cached file.
		target, err := filepath.EvalSymlinks(filepath.Join(cachePath, name))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("resolve cached manifest entry %q: %w", name, err)
		}
		if relative, err := filepath.Rel(cachePath, target); err == nil && filepath.IsLocal(relative) {
			protected[strings.ToLower(relative)] = struct{}{}
		}
	}
	return protected, nil
}

func findUnusedCacheFiles(ctx context.Context, cache fs.FS, protected map[string]struct{}) ([]string, error) {
	var candidates []string
	err := fs.WalkDir(cache, ".", func(name string, entry fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		dir, _, inCache := strings.Cut(name, "/")
		if !inCache {
			if entry.IsDir() && dir != cc.PackageCacheDir && dir != cc.ImageCacheDir {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		// The image cache also contains metadata such as images.node.mf.
		if dir == cc.ImageCacheDir && !images.HasSuffixI(entry.Name(), ".tar.gz", ".tgz", ".tar") {
			return nil
		}
		localPath := filepath.FromSlash(name)
		if _, ok := protected[strings.ToLower(localPath)]; !ok {
			candidates = append(candidates, localPath)
		}
		return nil
	})
	return candidates, err
}

func readCacheManifest(root *os.Root, versionDir string) ([]string, error) {
	name := filepath.Join("versions", versionDir, "installation.manifest")
	file, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()

	var paths []string
	scanner := bufio.NewScanner(file)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		item, err := manifest.ReadItem(line)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", name, lineNum, err)
		}
		// Manifests can reference downloads beyond the directories this command sweeps.
		if !fs.ValidPath(item.Path) || strings.ContainsAny(item.Path, "\\:\x00") ||
			item.Filename == "" || item.Filename == "." || item.Filename == ".." || strings.ContainsAny(item.Filename, "/\\:\x00") {
			return nil, fmt.Errorf("%s line %d: invalid cache path %q / %q", name, lineNum, item.Path, item.Filename)
		}
		paths = append(paths, item.FilePath(""))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("%s has no cache entries", name)
	}
	return paths, nil
}

func regularCacheFile(root *os.Root, name string) (fs.FileInfo, error) {
	// Check parents from the cache root down so aliases cannot select a file
	// through a symlinked cache directory.
	var dir string
	for _, part := range strings.Split(filepath.Dir(name), string(filepath.Separator)) {
		dir = filepath.Join(dir, part)
		info, err := root.Lstat(dir)
		if err != nil || !info.IsDir() {
			return nil, err
		}
	}
	info, err := root.Lstat(name)
	if err != nil || !info.Mode().IsRegular() {
		return nil, err
	}
	return info, nil
}
