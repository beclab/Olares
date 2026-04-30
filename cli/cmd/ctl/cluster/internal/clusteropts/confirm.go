package clusteropts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ConfirmDestructive guards destructive cluster verbs behind a y/N
// prompt unless assumeYes is set. Mirrors the well-tested
// confirmDestructive in cli/cmd/ctl/settings/vpn/common.go (the most
// complete copy in the project — TTY check + assumeYes short-circuit
// + literal y/yes match) so cluster mutating verbs share the same
// UX as the settings tree.
//
// Lifted into clusteropts (not a project-wide utility) so the cluster
// tree owns its own copy — same pattern the existing settings copies
// follow. If a future need lands to share it across umbrellas, the
// canonical location to lift it to would be a new pkg/cliutil/...
// package, not back into clusteropts.
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
