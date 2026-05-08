// Package cliutil hosts CLI-tree-agnostic helpers shared across the
// cmd/ctl/* subtrees (settings, cluster, files, dashboard, market).
// Each helper here is the single source of truth for a concern that
// would otherwise drift across copies — e.g. destructive-verb
// confirmation, output formatting, etc.
//
// Helpers in this package are intentionally minimal: no flags wired,
// no Cobra dependency unless absolutely necessary, no implicit
// stdout/stderr — callers pass io.Reader / io.Writer so tests can
// pin them. That keeps the import graph one-way: cmd/ctl/* → cliutil.
package cliutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ConfirmDestructive guards destructive verbs (delete, restart, scale
// to zero, suspend, ...) behind a y/N prompt unless assumeYes is set.
// Single source for the pattern the cluster + settings trees use; if
// you find yourself writing another local copy, import this one.
//
// Calling convention:
//
//   - prompt: where to write the "Are you sure?" line (typically
//     os.Stderr so JSON consumers on stdout don't see it).
//   - in: where to read the answer from (typically os.Stdin).
//   - message: the action being confirmed; rendered verbatim before
//     the [y/N]: suffix.
//   - assumeYes: if true, skip the prompt entirely. Map this to a
//     --yes / -y flag on the calling command.
//
// Behavior:
//
//   - assumeYes=true        → returns nil immediately.
//   - in is a non-TTY file  → returns an error rather than silently
//     proceeding (we'd rather break a script than delete something
//     the operator didn't review).
//   - answer in {y, yes}    → returns nil.
//   - any other answer / EOF → returns "aborted by user".
func ConfirmDestructive(prompt io.Writer, in io.Reader, message string, assumeYes bool) error {
	if assumeYes {
		return nil
	}
	if f, ok := in.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			return fmt.Errorf("stdin is not a terminal — pass --yes to confirm: %s", message)
		}
	}
	if _, err := fmt.Fprintf(prompt, "%s [y/N]: ", message); err != nil {
		return err
	}
	rd := bufio.NewReader(in)
	line, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted by user")
	}
}
