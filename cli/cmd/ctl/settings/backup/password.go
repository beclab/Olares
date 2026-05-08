package backup

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings backup password ...`
//
// The repository encryption password every backup plan binds to. Lives
// at user-service /api/backup/password/<name> (NOT the BFL backup-server
// /apis/backup/v1/...). The SPA's createBackupPlan flow writes it
// before POSTing the plan body, so the plan can refer to a password
// that exists at create time.
//
//	PUT /api/backup/password/{name}     body {password}
//
// The SPA does NOT expose a separate "change password later" UI — that
// flow is implicit (you'd recreate the plan). user-service does still
// accept PUTs to an existing record, so a CLI user can rotate the
// password independently of the SPA flow if needed.
//
// We deliberately don't ship a `password get` because the upstream
// shouldn't be returning the password in cleartext anyway, and
// surfacing whatever placeholder it does return would invite
// confusion.

func NewPasswordCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "backup repository password (Settings -> Backup, repo password)",
		Long: `Manage the encryption password the BFL backup-server stores per plan
name (separate from the plan record itself).

Subcommands:
  set <name>   create or update the repository password
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPasswordSetCommand(f))
	return cmd
}

func newPasswordSetCommand(f *cmdutil.Factory) *cobra.Command {
	var passwordStdin bool
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "create or update the repository password for a backup plan name",
		Long: `Set (or update) the repository encryption password bound to a backup
plan name. By default the password is read from a TTY without echo;
pass --password-stdin to read it once from stdin (useful in scripts).

Examples:
  olares-cli settings backup password set my-plan
  echo -n "my-secret-password" | olares-cli settings backup password set my-plan --password-stdin

The password is encrypted at rest in the backup-server repository; the
SPA's createBackupPlan flow uses it as the restic / kopia password
when sealing snapshots. Losing it means losing the ability to decrypt
existing snapshots — the upstream cannot recover this password.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPasswordSet(c.Context(), f, args[0], passwordStdin)
		},
	}
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin once (no prompt)")
	return cmd
}

func runPasswordSet(ctx context.Context, f *cmdutil.Factory, planName string, passwordStdin bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	planName = strings.TrimSpace(planName)
	if planName == "" {
		return fmt.Errorf("set requires a plan name")
	}
	password, err := readBackupPassword(passwordStdin)
	if err != nil {
		return err
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/api/backup/password/" + url.PathEscape(planName)
	body := map[string]string{"password": password}
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", path, body, nil); err != nil {
		return err
	}
	fmt.Printf("Set repository password for backup plan name %q.\n", planName)
	return nil
}

// readBackupPassword either reads stdin once (when --password-stdin is
// passed) or prompts twice on a TTY (input + confirmation, both with
// echo off) and rejects mismatches. Mirrors the same shape as
// settings/me/password.go's interactive flow.
func readBackupPassword(fromStdin bool) (string, error) {
	if fromStdin {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read --password-stdin: %w", err)
		}
		return strings.TrimRight(string(buf), "\n\r"), nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("stdin is not a terminal — pass --password-stdin to provide the password")
	}
	first, err := promptHidden("New backup password: ")
	if err != nil {
		return "", err
	}
	second, err := promptHidden("Re-enter password: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("passwords do not match")
	}
	return first, nil
}

func promptHidden(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", err
	}
	defer fmt.Fprintln(os.Stderr)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pw), nil
}
