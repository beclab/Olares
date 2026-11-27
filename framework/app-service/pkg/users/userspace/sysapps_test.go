package userspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserChartRootPinsPublishedBundle(t *testing.T) {
	root := t.TempDir()
	assertRoot := func(want string) {
		t.Helper()
		if got, err := userChartRoot(root); got != want || err != nil {
			t.Fatalf("user chart root = %q, %v; want %q", got, err, want)
		}
	}
	assertRoot(root)
	for _, name := range []string{".bundle.aaaaaaaa", ".bundle.bbbbbbbb"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	publish := func(target string) {
		t.Helper()
		stage := filepath.Join(root, ".next")
		if err := os.Symlink(target, stage); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(stage, filepath.Join(root, "current")); err != nil {
			t.Fatal(err)
		}
	}
	publish(".bundle.aaaaaaaa")
	oldRoot, err := userChartRoot(root)
	if err != nil || oldRoot != filepath.Join(root, ".bundle.aaaaaaaa") {
		t.Fatalf("unexpected initial bundle %q, %v", oldRoot, err)
	}
	publish(".bundle.bbbbbbbb")
	assertRoot(filepath.Join(root, ".bundle.bbbbbbbb"))
	if _, err := os.Stat(oldRoot); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"../.bundle.aaaaaaaa", ".bundle.cccccccc"} {
		publish(bad)
		if _, err := userChartRoot(root); err == nil {
			t.Fatalf("accepted invalid or missing snapshot %q", bad)
		}
	}
}
