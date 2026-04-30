package upload

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// writeFile is a tiny helper for the walker tests so each test reads
// linearly without inline `os.Create` / `defer Close` clutter.
func writeFile(t *testing.T, p string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBuildPlan_SingleFile_RemoteIsDir: trailing slash on remote means
// "upload into this directory"; the server-side filename is the local
// basename.
func TestBuildPlan_SingleFile_RemoteIsDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	writeFile(t, src, "hello")
	plan, err := BuildPlan(src, "Documents/", "/drive/Home", "/drive/Home")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/drive/Home/Documents/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
	}
	if plan.ChunkParentDir != "/drive/Home/Documents/" {
		t.Errorf("ChunkParentDir = %q", plan.ChunkParentDir)
	}
	if plan.RelativeRoot != "Documents" {
		t.Errorf("RelativeRoot = %q", plan.RelativeRoot)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("Files = %d", len(plan.Files))
	}
	f := plan.Files[0]
	if f.RelativePath != "report.pdf" || f.RemoteName != "report.pdf" {
		t.Errorf("file = %+v", f)
	}
}

// TestBuildPlan_SingleFile_RenameOnUpload: no trailing slash means
// the remote path IS the destination; we split it into parent + base.
func TestBuildPlan_SingleFile_RenameOnUpload(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	writeFile(t, src, "x")
	plan, err := BuildPlan(src, "Documents/2026.pdf", "/drive/Home", "/drive/Home")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/drive/Home/Documents/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
	}
	if plan.RelativeRoot != "Documents" {
		t.Errorf("RelativeRoot = %q", plan.RelativeRoot)
	}
	f := plan.Files[0]
	if f.RemoteName != "2026.pdf" || f.RelativePath != "2026.pdf" {
		t.Errorf("file = %+v", f)
	}
}

// TestBuildPlan_DirectoryRequiresTrailingSlash: lowest-friction error
// when user forgets the trailing slash on a directory destination.
func TestBuildPlan_DirectoryRequiresTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	if _, err := BuildPlan(dir, "Backups", "/drive/Home", "/drive/Home"); err == nil {
		t.Fatal("expected error when local is dir but remote has no trailing slash")
	}
}

// TestBuildPlan_Directory_Recursion: walks a small tree and checks the
// emitted file list + the empty-dir mkdir list. Source basename
// becomes the top-level component under remote.
func TestBuildPlan_Directory_Recursion(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "mydir")
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	writeFile(t, filepath.Join(src, "sub", "b.txt"), "bb")
	writeFile(t, filepath.Join(src, "sub", "deep", "c.txt"), "ccc")
	if err := os.MkdirAll(filepath.Join(src, "empty", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(src, "Backups/", "/drive/Home", "/drive/Home")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/drive/Home/Backups/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
	}
	if plan.ChunkParentDir != "/drive/Home/Backups/" {
		t.Errorf("ChunkParentDir = %q", plan.ChunkParentDir)
	}

	gotFiles := make([]string, 0, len(plan.Files))
	for _, f := range plan.Files {
		gotFiles = append(gotFiles, f.RelativePath)
	}
	wantFiles := []string{
		"mydir/a.txt",
		"mydir/sub/b.txt",
		"mydir/sub/deep/c.txt",
	}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Errorf("files:\n got  %v\n want %v", gotFiles, wantFiles)
	}

	// Empty dir + its (also empty) child should be in EmptyDirs.
	gotEmpty := append([]string(nil), plan.EmptyDirs...)
	sort.Strings(gotEmpty)
	wantEmpty := []string{"mydir/empty", "mydir/empty/nested"}
	if !reflect.DeepEqual(gotEmpty, wantEmpty) {
		t.Errorf("empty dirs:\n got  %v\n want %v", gotEmpty, wantEmpty)
	}
}

// TestBuildPlan_SyncRoot: Sync uploads keep the `/sync/<repo>/<sub>/`
// shape for the API queries (upload-link, file-uploaded-bytes) but
// the chunk POST's parent_dir form field MUST be the path INSIDE the
// repo (`/<sub>/`) so Seafile's seafhttp/upload-aj endpoint can
// resolve it relative to the token-pinned repo root.
func TestBuildPlan_SyncRoot(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	writeFile(t, src, "hello")
	plan, err := BuildPlan(src, "docs/", "/sync/repo-123", "")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/sync/repo-123/docs/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
	}
	if plan.ChunkParentDir != "/docs/" {
		t.Errorf("ChunkParentDir = %q, want %q", plan.ChunkParentDir, "/docs/")
	}
	if plan.RelativeRoot != "docs" {
		t.Errorf("RelativeRoot = %q", plan.RelativeRoot)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("Files = %d", len(plan.Files))
	}
}

// TestBuildPlan_SyncRepoRoot: uploading directly to a Sync repo root
// (no <sub> path) yields ChunkParentDir == "/" — the Seafile token
// pins the repo, parent_dir tells the server where INSIDE that repo
// to land the file, and "/" means "repo root".
func TestBuildPlan_SyncRepoRoot(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	writeFile(t, src, "hello")
	plan, err := BuildPlan(src, "", "/sync/repo-123", "")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/sync/repo-123/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
	}
	if plan.ChunkParentDir != "/" {
		t.Errorf("ChunkParentDir = %q, want %q", plan.ChunkParentDir, "/")
	}
}

// TestBuildPlan_NamespaceShapes covers the per-namespace
// (apiRoot, chunkRoot) → (ParentDir, ChunkParentDir) mapping
// for the namespaces that share the files-backend protocol
// (Drive/Data/Cache/External all have apiRoot == chunkRoot;
// only Sync needs the inside-repo split, which is exercised
// in TestBuildPlan_SyncRoot above).
//
// Keeping the cases table-driven means adding a new namespace to
// uploadRootAndDriveType doesn't require copy-pasting another whole
// test, just a row here.
func TestBuildPlan_NamespaceShapes(t *testing.T) {
	cases := []struct {
		name                 string
		remoteSub            string
		apiRoot              string
		chunkRoot            string
		wantParentDir        string
		wantChunkParentDir   string
		wantRelativeRoot     string
	}{
		{
			name:               "drive Data subdir",
			remoteSub:          "Backups/",
			apiRoot:            "/drive/Data",
			chunkRoot:          "/drive/Data",
			wantParentDir:      "/drive/Data/Backups/",
			wantChunkParentDir: "/drive/Data/Backups/",
			wantRelativeRoot:   "Backups",
		},
		{
			name:               "drive Data root",
			remoteSub:          "",
			apiRoot:            "/drive/Data",
			chunkRoot:          "/drive/Data",
			wantParentDir:      "/drive/Data/",
			wantChunkParentDir: "/drive/Data/",
			wantRelativeRoot:   "",
		},
		{
			name:               "cache app subdir",
			remoteSub:          "AppName/data/",
			apiRoot:            "/cache/node-1",
			chunkRoot:          "/cache/node-1",
			wantParentDir:      "/cache/node-1/AppName/data/",
			wantChunkParentDir: "/cache/node-1/AppName/data/",
			wantRelativeRoot:   "AppName/data",
		},
		{
			name:               "external volume subdir",
			remoteSub:          "hdd1/Movies/",
			apiRoot:            "/external/node-1",
			chunkRoot:          "/external/node-1",
			wantParentDir:      "/external/node-1/hdd1/Movies/",
			wantChunkParentDir: "/external/node-1/hdd1/Movies/",
			wantRelativeRoot:   "hdd1/Movies",
		},
		// Cloud-drive shapes — same chunkRoot==apiRoot invariant,
		// just different prefix and SubPath conventions (awss3
		// surfaces <bucket>/<key> as the SubPath; google/dropbox
		// have no <bucket>-equivalent so SubPath starts at the
		// account root).
		{
			name:               "awss3 bucket subdir",
			remoteSub:          "bucket/Backups/",
			apiRoot:            "/awss3/account-x",
			chunkRoot:          "/awss3/account-x",
			wantParentDir:      "/awss3/account-x/bucket/Backups/",
			wantChunkParentDir: "/awss3/account-x/bucket/Backups/",
			wantRelativeRoot:   "bucket/Backups",
		},
		{
			name:               "google account subdir",
			remoteSub:          "Documents/",
			apiRoot:            "/google/account-x",
			chunkRoot:          "/google/account-x",
			wantParentDir:      "/google/account-x/Documents/",
			wantChunkParentDir: "/google/account-x/Documents/",
			wantRelativeRoot:   "Documents",
		},
		{
			name:               "dropbox account subdir",
			remoteSub:          "Notes/",
			apiRoot:            "/dropbox/account-x",
			chunkRoot:          "/dropbox/account-x",
			wantParentDir:      "/dropbox/account-x/Notes/",
			wantChunkParentDir: "/dropbox/account-x/Notes/",
			wantRelativeRoot:   "Notes",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "report.pdf")
			writeFile(t, src, "hello")
			plan, err := BuildPlan(src, tc.remoteSub, tc.apiRoot, tc.chunkRoot)
			if err != nil {
				t.Fatal(err)
			}
			if plan.ParentDir != tc.wantParentDir {
				t.Errorf("ParentDir = %q, want %q", plan.ParentDir, tc.wantParentDir)
			}
			if plan.ChunkParentDir != tc.wantChunkParentDir {
				t.Errorf("ChunkParentDir = %q, want %q",
					plan.ChunkParentDir, tc.wantChunkParentDir)
			}
			if plan.RelativeRoot != tc.wantRelativeRoot {
				t.Errorf("RelativeRoot = %q, want %q",
					plan.RelativeRoot, tc.wantRelativeRoot)
			}
			if len(plan.Files) != 1 {
				t.Fatalf("Files = %d", len(plan.Files))
			}
		})
	}
}
