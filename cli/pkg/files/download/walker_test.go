package download

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
)

// fakeTree is a trivial in-memory remote filesystem used to drive the
// walker via httptest. Keys are full plain paths (e.g.
// "drive/Home/Documents") to a `dirContents` value (nil for files).
type fakeTree map[string]*dirContents

type dirContents struct {
	entries []fakeEntry
}

type fakeEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// handler exposes the tree as the subset of /api/resources the walker
// touches: listing requests (path ends in '/') return the dirContents
// for the matching key. Stat requests aren't used by the walker
// (BuildPlan only calls List), so we don't need to fake those.
func (t fakeTree) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /api/resources/<plain> — strip the prefix and percent-decode
		// each segment to get the lookup key.
		raw := strings.TrimPrefix(r.URL.Path, "/api/resources/")
		hadTrailing := strings.HasSuffix(raw, "/")
		raw = strings.Trim(raw, "/")
		segs := strings.Split(raw, "/")
		for i, s := range segs {
			d, err := url.PathUnescape(s)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			segs[i] = d
		}
		key := strings.Join(segs, "/")
		_ = hadTrailing
		dc, ok := t[key]
		if !ok || dc == nil {
			http.Error(w, "not a directory", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"items": dc.entries,
		})
	})
}

func TestBuildPlan_FlatDirectory(t *testing.T) {
	tree := fakeTree{
		"drive/Home/Documents": {
			entries: []fakeEntry{
				{Name: "a.txt", Size: 10},
				{Name: "b.txt", Size: 20},
			},
		},
	}
	client, _ := newTestClient(t, tree.handler())

	plan, err := BuildPlan(context.Background(), client, "drive/Home/Documents", "/tmp/dest")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(plan.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(plan.Files))
	}
	if !strings.HasSuffix(plan.LocalRoot, "Documents") {
		t.Errorf("LocalRoot should preserve remote leaf name, got %q", plan.LocalRoot)
	}
	if plan.Files[0].RelativePath != "a.txt" || plan.Files[1].RelativePath != "b.txt" {
		t.Errorf("RelativePaths: %v", []string{plan.Files[0].RelativePath, plan.Files[1].RelativePath})
	}
	for _, f := range plan.Files {
		if !strings.HasPrefix(f.RemotePlainPath, "drive/Home/Documents/") {
			t.Errorf("RemotePlainPath should be drive/Home/Documents/<name>, got %q", f.RemotePlainPath)
		}
	}
}

func TestBuildPlan_NestedWithEmptyDir(t *testing.T) {
	tree := fakeTree{
		"drive/Home/Backups": {
			entries: []fakeEntry{
				{Name: "top.txt", Size: 5},
				{Name: "photos", IsDir: true},
				{Name: "empty", IsDir: true},
			},
		},
		"drive/Home/Backups/photos": {
			entries: []fakeEntry{
				{Name: "img.jpg", Size: 100},
			},
		},
		"drive/Home/Backups/empty": {
			entries: nil, // genuinely empty subdir
		},
	}
	client, _ := newTestClient(t, tree.handler())

	plan, err := BuildPlan(context.Background(), client, "drive/Home/Backups", "/tmp/out")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	// Files: 2 ("top.txt" and "photos/img.jpg").
	if len(plan.Files) != 2 {
		t.Fatalf("want 2 files, got %d (%+v)", len(plan.Files), plan.Files)
	}
	relPaths := []string{plan.Files[0].RelativePath, plan.Files[1].RelativePath}
	want := []string{"photos/img.jpg", "top.txt"}
	for i, w := range want {
		if relPaths[i] != w {
			t.Errorf("RelativePath[%d]: want %q, got %q", i, w, relPaths[i])
		}
	}
	// Empty subdir is captured.
	if len(plan.EmptyDirs) != 1 || plan.EmptyDirs[0] != "empty" {
		t.Errorf("EmptyDirs: want [empty], got %v", plan.EmptyDirs)
	}
}

func TestBuildPlan_DepthFirstOrdering(t *testing.T) {
	// 3-deep tree — verify shallow-first sort on EmptyDirs and
	// deterministic Files ordering. The walker is depth-first
	// internally but Plan sorts before returning.
	tree := fakeTree{
		"r": {entries: []fakeEntry{{Name: "deep", IsDir: true}, {Name: "leaf", Size: 1}}},
		"r/deep": {entries: []fakeEntry{
			{Name: "x", IsDir: true},
			{Name: "y.txt", Size: 1},
		}},
		"r/deep/x": {entries: nil}, // empty
	}
	client, _ := newTestClient(t, tree.handler())
	plan, err := BuildPlan(context.Background(), client, "r", "/tmp")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if got := []string{plan.Files[0].RelativePath, plan.Files[1].RelativePath}; !equal(got, []string{"deep/y.txt", "leaf"}) {
		t.Errorf("Files ordering: got %v", got)
	}
	if !equal(plan.EmptyDirs, []string{"deep/x"}) {
		t.Errorf("EmptyDirs: got %v", plan.EmptyDirs)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Sanity check: the tree handler should at least pretend to URL-encode,
// so this no-op test confirms no panics on funky names. We don't go
// deeper because the actual encoding lives in upload.EncodeURL, which
// has its own test suite.
func TestList_PathEncoding_Roundtrip(t *testing.T) {
	want := "Special !'()*"
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode URL — make sure the un-encoded segment matches what we asked for.
		raw := strings.TrimPrefix(r.URL.Path, "/api/resources/")
		raw = strings.Trim(raw, "/")
		segs := strings.Split(raw, "/")
		decoded := make([]string, len(segs))
		for i, s := range segs {
			d, err := url.PathUnescape(s)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			decoded[i] = d
		}
		got := path.Join(decoded...)
		if got != want {
			t.Errorf("decoded path: want %q, got %q", want, got)
		}
		_, _ = io.WriteString(w, `{"items":[]}`)
	}))
	if _, err := client.List(context.Background(), want); err != nil {
		t.Fatalf("List: %v", err)
	}
}
