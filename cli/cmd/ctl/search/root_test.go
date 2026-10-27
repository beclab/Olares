package search

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestSearchCommandExposesFederatedDriveOnly(t *testing.T) {
	cmd := NewSearchCommand(&cmdutil.Factory{})
	registered := make(map[string]bool)
	for _, child := range cmd.Commands() {
		registered[child.Name()] = true
	}

	if !registered["drive"] {
		t.Fatal("drive command is not registered")
	}
	drive, _, err := cmd.Find([]string{"drive"})
	if err != nil {
		t.Fatal(err)
	}
	if drive.Flags().Lookup("watch") == nil {
		t.Fatal("drive command does not expose --watch")
	}
	for _, removed := range []string{"gdrive", "dropbox"} {
		if registered[removed] {
			t.Fatalf("removed command %q is still registered", removed)
		}
	}
}

// Only the commands that can run the federated job may default to printing
// everything; `knowledge` never leaves the session API, which pages
// server-side and has no completed result set to print in full.
func TestSearchDefaultLimitPerCommand(t *testing.T) {
	cmd := NewSearchCommand(&cmdutil.Factory{})
	cases := map[string]string{
		"drive":     "0",
		"sync":      "0",
		"knowledge": "20",
	}
	for name, want := range cases {
		found, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatalf("find %s: %v", name, err)
		}
		flag := found.Flags().Lookup("limit")
		if flag == nil {
			t.Fatalf("%s has no --limit", name)
		}
		if flag.DefValue != want {
			t.Fatalf("%s --limit default = %s, want %s", name, flag.DefValue, want)
		}
	}
}

func TestPagingOptionsSessionFallback(t *testing.T) {
	t.Parallel()

	// An unlimited request cannot be served by the session API, which pages
	// blind over a server-side cache; it falls back to that page.
	unlimited := pagingOptions{limit: 0, offset: 0}
	if got := unlimited.sessionPaging().limit; got != initPageSize {
		t.Fatalf("sessionPaging().limit = %d, want %d", got, initPageSize)
	}
	if !unlimited.unlimited() || unlimited.windowed() {
		t.Fatal("an unlimited request must not look windowed")
	}

	// An explicit window is the user's, and survives untouched.
	explicit := pagingOptions{limit: 50, offset: 20}
	resolved := explicit.sessionPaging()
	if resolved.limit != 50 || resolved.offset != 20 {
		t.Fatalf("sessionPaging() = %+v, want the request unchanged", resolved)
	}
	if !resolved.windowed() {
		t.Fatal("an explicit window must look windowed")
	}
}

func TestPagingOptionsValidateAcceptsUnlimited(t *testing.T) {
	t.Parallel()

	if _, err := (&pagingOptions{limit: 0, output: "table"}).validate(); err != nil {
		t.Fatalf("--limit 0 must be accepted: %v", err)
	}
	if _, err := (&pagingOptions{limit: -1, output: "table"}).validate(); err == nil {
		t.Fatal("--limit -1 must be rejected")
	}
}

// The Desktop dialog labels the single federated entry "Files", so that name
// has to reach the same command as `drive` rather than 404 into usage output.
func TestSearchFilesAliasResolvesToDrive(t *testing.T) {
	cmd := NewSearchCommand(&cmdutil.Factory{})
	resolved, args, err := cmd.Find([]string{"files"})
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("unresolved command tokens: %v", args)
	}
	if resolved.Name() != "drive" {
		t.Fatalf("search files resolved to %q, want drive", resolved.Name())
	}
}
