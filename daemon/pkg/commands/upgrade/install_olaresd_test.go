package upgrade

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

func TestOldPrepareLogCannotSkipDaemonReplacement(t *testing.T) {
	base, binary := t.TempDir(), filepath.Join(t.TempDir(), "olares-cli")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho 'Error: new prepare failed' >&2\nexit 7\n"), 0755); err != nil {
		t.Fatal(err)
	}
	logs := filepath.Join(base, "versions", "v1.12.8", "logs")
	if err := os.MkdirAll(logs, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logs, "install.log"), []byte("UpdateOlaresdEnv success\n"), 0644); err != nil {
		t.Fatal(err)
	}
	oldBase := commands.TERMINUS_BASE_DIR
	commands.TERMINUS_BASE_DIR = base
	t.Cleanup(func() { commands.TERMINUS_BASE_DIR = oldBase })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	prepare := NewInstallOlaresd().(*prepareOlaresd)
	prepare.cliPath = binary
	res, err := prepare.Execute(ctx, state.UpgradeTarget{Version: *semver.MustParse("1.12.8")})
	if err != nil {
		t.Fatal(err)
	}
	execution := res.(ExecutionRes)
	if execution.Finished() {
		t.Fatal("old success log skipped daemon replacement")
	}
	err = AwaitExecution(ctx, execution, nil)
	for range execution.Progress() {
	} // Wait for the log reader before TempDir cleanup.
	if err == nil || !strings.Contains(err.Error(), "new prepare failed") {
		t.Fatalf("error = %v", err)
	}
}
