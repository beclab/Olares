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
