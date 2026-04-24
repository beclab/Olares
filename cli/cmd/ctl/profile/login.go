package profile

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

type loginOptions struct {
	commonCredFlags
	passwordStdin bool
	totp          string
	noSwitch      bool
}

// NewLoginCommand: `olares-cli profile login --olares-id <id> [...]`
//
// Mode A (password login). Behavior matrix from the design doc:
//   - profile does not exist            → auto-create (with provided overrides)
//   - profile exists, no/expired token  → reuse existing profile, write new token
//   - profile exists, valid token       → reject with `profile remove` hint
//
// Password is read from stdin when --password-stdin is set, otherwise from
// the controlling terminal (with input echoing disabled). Two-factor accounts
// must supply --totp.
func NewLoginCommand() *cobra.Command {
	o := &loginOptions{}
	cmd := &cobra.Command{
		Use:   "login",
		Short: "log in to an Olares instance with a password (mode A)",
		Long: `Authenticate to an Olares instance using the user's password (and TOTP if 2FA is enabled).

The profile is auto-created on first login. Re-running login against an
already-authenticated profile is rejected; remove the profile first
(` + "`olares-cli profile remove <id>`" + `) and log in again.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLogin(cmd.Context(), o)
		},
	}
	o.commonCredFlags.bind(cmd)
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false, "read the password from stdin instead of prompting")
	cmd.Flags().StringVar(&o.totp, "totp", "", "TOTP code for accounts with two-factor authentication enabled")
	cmd.Flags().BoolVar(&o.noSwitch, "no-switch", false, "do not change the current profile after a successful login (useful for scripts)")
	return cmd
}

// bind wires the cred flags onto a cobra.Command. Defined here (not on
// commonCredFlags directly) to keep the import-side flag set identical.
func (f *commonCredFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.olaresID, "olares-id", "", "olaresId, e.g. alice@olares.com (required)")
	cmd.Flags().StringVar(&f.name, "name", "", "optional alias for the profile (defaults to the olaresId)")
	cmd.Flags().StringVar(&f.authURLOverride, "auth-url-override", "", "override the derived auth URL (dev/internal use)")
	cmd.Flags().StringVar(&f.localURLPrefix, "local-url-prefix", "", "label inserted between the auth subdomain and the terminus name (dev/internal use)")
	cmd.Flags().BoolVar(&f.insecureSkipVerify, "insecure-skip-verify", false, "disable TLS verification for HTTP calls under this profile (dev/internal use)")
}

func runLogin(ctx context.Context, o *loginOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	id, terminusName, authURL, err := o.commonCredFlags.validateAndDeriveAuthURL()
	if err != nil {
		return err
	}

	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	store, err := auth.NewFileStore()
	if err != nil {
		return err
	}
	profile, err := ensureProfileWritable(cfg, store, o.commonCredFlags, time.Now())
	if err != nil {
		return err
	}

	password, err := readPassword(o.passwordStdin, o.olaresID)
	if err != nil {
		return err
	}

	tok, err := loginWithTOTPPrompt(ctx, auth.LoginRequest{
		AuthURL:            authURL,
		LocalName:          id.Local(),
		TerminusName:       terminusName,
		Password:           password,
		TOTP:               o.totp,
		InsecureSkipVerify: o.insecureSkipVerify,
	}, o.olaresID)
	if err != nil {
		return err
	}

	res, err := persistTokenAndProfile(cfg, store, profile, tok, !o.noSwitch)
	if err != nil {
		return err
	}

	fmt.Printf("logged in as %s (profile: %s)\n", o.olaresID, profile.DisplayName())
	printSwitchNotice(res, profile.DisplayName())
	printPlaintextWarning()
	return nil
}

// loginWithTOTPPrompt wraps auth.Login with one round of interactive TOTP
// recovery: if the first attempt comes back ErrTOTPRequired (meaning the
// account has 2FA enabled and the caller didn't supply --totp) AND we're
// running on a TTY, prompt the user for the 6-digit code and retry once.
//
// If --totp was already supplied OR stdin is not a TTY (e.g. piped via
// --password-stdin from a script), we degrade to the original error so the
// caller knows to re-run with --totp explicitly.
//
// Note the retry re-issues /api/firstfactor — the server doesn't keep
// transitional state between the two factor steps from our perspective, and
// re-validating the password is cheap. We do NOT re-prompt the password.
func loginWithTOTPPrompt(ctx context.Context, req auth.LoginRequest, olaresID string) (*auth.Token, error) {
	tok, err := auth.Login(ctx, req)
	if err == nil {
		return tok, nil
	}
	if !errors.Is(err, auth.ErrTOTPRequired) || req.TOTP != "" {
		return nil, err
	}
	if !term.IsTerminal(int(syscall.Stdin)) {
		return nil, fmt.Errorf("two-factor authentication required: re-run with --totp <code>")
	}
	totp, perr := promptTOTP(olaresID)
	if perr != nil {
		return nil, perr
	}
	req.TOTP = totp
	return auth.Login(ctx, req)
}

// promptTOTP reads a 6-digit code from the controlling terminal. The code is
// short-lived and not secret-sensitive in the same way a password is, so we
// echo it (matches `gh auth login`, `aws sso login`, kubectl OIDC plugins).
func promptTOTP(olaresID string) (string, error) {
	fmt.Printf("two-factor code for %s: ", olaresID)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read TOTP: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", errors.New("TOTP code is empty")
	}
	return line, nil
}

// readPassword pulls the password from the requested source. --password-stdin
// reads exactly one line from stdin (newline stripped); the interactive path
// turns off terminal echo. We never log or print the password.
func readPassword(fromStdin bool, olaresID string) (string, error) {
	if fromStdin {
		return readSingleLine(os.Stdin)
	}
	if !term.IsTerminal(int(syscall.Stdin)) {
		return "", errors.New("stdin is not a terminal; pass --password-stdin and pipe the password instead")
	}
	fmt.Printf("password for %s: ", olaresID)
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if len(pw) == 0 {
		return "", errors.New("password is empty")
	}
	return string(pw), nil
}

// readSingleLine reads up to and including the first '\n' (or EOF) from r and
// returns the trimmed line. Used for --password-stdin so that
// `printf '%s' "$P" | olares-cli profile login --password-stdin` works
// regardless of whether the input has a trailing newline.
func readSingleLine(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return "", errors.New("password is empty")
	}
	return line, nil
}
