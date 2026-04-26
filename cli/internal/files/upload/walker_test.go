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
	plan, err := BuildPlan(src, "Documents/")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/drive/Home/Documents/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
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
	plan, err := BuildPlan(src, "Documents/2026.pdf")
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
	if _, err := BuildPlan(dir, "Backups"); err == nil {
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

	plan, err := BuildPlan(src, "Backups/")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ParentDir != "/drive/Home/Backups/" {
		t.Errorf("ParentDir = %q", plan.ParentDir)
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
