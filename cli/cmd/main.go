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
	"github.com/beclab/Olares/cli/pkg/cmdutil"
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

	machineReadable := cmdutil.AskedForJSON(executed)

	if shouldAnnounceSkillDrift(machineReadable, os.Args) {
		skills.Notice(os.Stderr, version.VERSION)
	}

	if err != nil {
		if errors.Is(err, clierr.ErrAlreadyReported) {
			os.Exit(1)
		}
		if !machineReadable || !clierr.WriteEnvelope(os.Stderr, err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

// shouldAnnounceSkillDrift decides whether this invocation is one the
// staleness notice belongs on.
//
// Skills installed on this machine outlive the binary that wrote them, so
// upgrading olares-cli leaves an agent reading instructions for a version it
// is not running. The check runs from main because this is the one place
// every invocation passes through: a PersistentPreRun on the root command is
// skipped by any subtree that declares one of its own.
//
// Two invocations are exempt. The `skills` tree is where the fix lives, and
// telling somebody running `skills install` that their skills are stale is
// noise. And anything that asked for JSON gets nothing: this is advisory
// prose, and it lands on the stream a failing `-o json` answers on, so a
// caller feeding stderr to a parser would get the notice and the error
// envelope concatenated, which parses as neither.
// skillDriftExempt are the verbs the staleness notice must not follow.
//
//   - skills: the verb that fixes it, which would otherwise report the
//     problem it has just resolved.
//   - update: the same, one level up — `update` re-runs `skills install` from
//     the binary it installed, so by the time this process exits the store is
//     correct and the only thing left that disagrees with it is this process,
//     which is the copy being replaced. Announcing drift there tells the user
//     their successful update failed.
//   - version: answers the same question in its own output, with both digests
//     and an in_sync field, rather than as a line of prose beside it.
var skillDriftExempt = map[string]bool{"skills": true, "update": true, "version": true}

func shouldAnnounceSkillDrift(machineReadable bool, args []string) bool {
	if machineReadable {
		return false
	}
	return len(args) < 2 || !skillDriftExempt[args[1]]
}
