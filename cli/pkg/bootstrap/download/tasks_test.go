package download

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/utils"
)

type cacheRuntime struct {
	connector.Runtime
	runner *connector.Runner
	linux  bool
}

func (r *cacheRuntime) GetRunner() *connector.Runner { return r.runner }
func (r *cacheRuntime) GetSystemInfo() connector.Systems {
	return &cacheSystem{linux: r.linux}
}

type cacheSystem struct {
	connector.Systems
	linux bool
}

func (s *cacheSystem) IsLinux() bool   { return s.linux }
func (*cacheSystem) IsWsl() bool       { return false }
func (*cacheSystem) GetOsArch() string { return "amd64" }

type cacheHost struct {
	connector.Host
	exec func(string) (string, error)
}

func (*cacheHost) SudoPrefixIfNecessary(cmd string) string { return cmd }
func (*cacheHost) FileExist(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
func (h *cacheHost) ExecExt(cmd string, _, _ bool) (string, int, error) {
	out, err := h.exec(cmd)
	if err != nil {
		return out, 1, err
	}
	return out, 0, nil
}

func TestCheckLocalCacheWithContainerd(t *testing.T) {
	const image = "docker.io/beclab/example:v1"
	tests := []struct {
		name        string
		output      string
		queryErr    error
		nonLinux    bool
		archive     string
		wantImages  []string
		wantQueries int
	}{
		{name: "missing archive with cached tag", output: image, wantImages: []string{image}, wantQueries: 1},
		{name: "different tag still downloads", output: "docker.io/beclab/example:v2", wantQueries: 1},
		{name: "different registry still downloads", output: "other.example/beclab/example:v1", wantQueries: 1},
		{name: "different repository still downloads", output: "docker.io/other/example:v1", wantQueries: 1},
		{name: "empty containerd cache", wantQueries: 1},
		{name: "containerd not installed", queryErr: errors.New("ctr not found"), wantQueries: 1},
		{name: "containerd stopped", queryErr: errors.New("socket missing"), wantQueries: 1},
		{name: "containerd query times out", output: image, queryErr: errors.New("context deadline exceeded"), wantQueries: 1},
		{name: "non Linux", nonLinux: true},
		{name: "valid archives need no containerd query", archive: "valid", wantImages: []string{image, "docker.io/beclab/second:v1"}},
		{name: "invalid archive with cached tag", archive: "corrupt", output: image, wantImages: []string{image}, wantQueries: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			var lines []string
			for _, item := range []struct{ filename, kind, id string }{
				{"first.tar", "images.test", "beclab/example:v1"},
				{"second.tar", "images.test", "docker.io/beclab/second:v1"},
				// Packages must still download even if their ID matches a cached image.
				{"package.tar", "binary", image},
			} {
				lines = append(lines, fmt.Sprintf("%s,cache,%s,/amd64,%s,/arm64,%s,%s", item.filename, item.kind, utils.MD5("valid"), utils.MD5("valid"), item.id))
				if tt.archive != "" && strings.HasPrefix(item.kind, "images.") {
					if err := os.MkdirAll(filepath.Join(baseDir, "cache"), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(baseDir, "cache", item.filename), []byte(tt.archive), 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			manifestPath := filepath.Join(baseDir, "manifest")
			if err := os.WriteFile(manifestPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
				t.Fatal(err)
			}
			queries := 0
			host := &cacheHost{exec: func(cmd string) (string, error) {
				queries++
				for _, required := range []string{"command -v ctr", "test -S /run/containerd/containerd.sock", "ctr --connect-timeout 5s --timeout 5s -n k8s.io images list -q"} {
					if !strings.Contains(cmd, required) {
						t.Errorf("query %q is missing %q", cmd, required)
					}
				}
				return tt.output, tt.queryErr
			}}
			runtime := &cacheRuntime{runner: &connector.Runner{Host: host}, linux: !tt.nonLinux}
			d := &PackageDownload{Manifest: manifestPath, BaseDir: baseDir}
			if err := d.CheckLocalCache(runtime); err != nil {
				t.Fatal(err)
			}
			var gotImages []string
			for _, item := range d.existingItems {
				gotImages = append(gotImages, normalizeImageRef(item.ImageName))
			}
			if !reflect.DeepEqual(gotImages, tt.wantImages) {
				t.Errorf("cached images = %v, want %v", gotImages, tt.wantImages)
			}
			if len(d.missingItems) != 3-len(tt.wantImages) {
				t.Errorf("missing items = %d, want %d", len(d.missingItems), 3-len(tt.wantImages))
			}
			if queries != tt.wantQueries {
				t.Errorf("containerd queries = %d, want %d", queries, tt.wantQueries)
			}
		})
	}
}

func TestNormalizeImageRef(t *testing.T) {
	for input, want := range map[string]string{
		"nginx:1.27":                   "docker.io/library/nginx:1.27",
		"docker.io/library/nginx:1.27": "docker.io/library/nginx:1.27",
		"beclab/example:v1":            "docker.io/beclab/example:v1",
		"invalid image":                "",
	} {
		if got := normalizeImageRef(input); got != want {
			t.Errorf("normalizeImageRef(%q) = %q, want %q", input, got, want)
		}
	}
}
