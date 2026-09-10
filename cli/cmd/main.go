package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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
	// ExecuteContextC hands back the command that actually ran, which is
	// how the failure below can tell whether this invocation asked to be
	// read by a program.
	executed, err := cmd.ExecuteContextC(ctx)

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
		if !askedForJSON(executed) || !clierr.WriteEnvelope(os.Stderr, err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

// askedForJSON reports whether the invocation that failed had asked for
// machine-readable output. Every tree spells that the same way, so one
// lookup covers all of them; a command without the flag is a human one
// and keeps the plain line.
func askedForJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		return false
	}
	return flag.Value.String() == "json"
}
