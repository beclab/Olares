package workload

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestWorkloadCommandRejectsRemovedUnusedImagesRoute(t *testing.T) {
	cmd := NewWorkloadCommand(cmdutil.NewFactory())
	cmd.SetArgs([]string{"unused-images"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected removed workload unused-images route to fail")
	}
}
