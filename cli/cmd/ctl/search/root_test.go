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
