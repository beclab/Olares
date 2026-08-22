package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/beclab/Olares/cli/cmd/ctl"
	"github.com/beclab/Olares/cli/cmd/ctl/skills"
	"github.com/beclab/Olares/cli/pkg/clierr"
	"github.com/beclab/Olares/cli/version"
)

func main() {
	// Install a SIGINT/SIGTERM handler that cancels ctx on the first
	// signal and hard-exits on the second. This is what the dead
	// pkg/signals.SetupSignalHandler used to do, now expressed with
	// stdlib primitives.
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		// Wait for a second signal -- if it arrives, exit immediately
		// instead of letting goroutines hang on cleanup.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Fprintln(os.Stderr, "received second signal, force exiting")
		os.Exit(1)
	}()

	cmd := ctl.NewDefaultCommand()
	err := cmd.ExecuteContext(ctx)

	// Skills installed on this machine outlive the binary that wrote them, so
	// upgrading olares-cli leaves an agent reading instructions for a version
	// it is not running. Said here because this is the one place every
	// invocation passes through: a PersistentPreRun on the root command is
	// skipped by any subtree that declares one of its own. Not said for the
	// `skills` tree itself, which is where the fix is.
	if len(os.Args) < 2 || os.Args[1] != "skills" {
		skills.Notice(os.Stderr, version.VERSION)
	}

	if err != nil {
		if errors.Is(err, clierr.ErrAlreadyReported) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
