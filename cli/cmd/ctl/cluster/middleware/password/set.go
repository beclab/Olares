package password

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// supportedTypes mirrors MiddlewareType in
// apps/.../controlPanelCommon/network/middleware.ts. We validate
// client-side so a typo in --type fails before the prompt rather
// than after (the operator types a password they can't undo).
var supportedTypes = []string{
	"mongodb", "postgres", "redis", "rabbitmq", "minio",
	"nats", "mysql", "mariadb", "elasticsearch",
}

// NewSetCommand: `olares-cli cluster middleware password set --type X
// --name N --namespace NS --user U [--password P] [--yes]`.
//
// Calls SPA's updateMiddlewarePassword
// (apps/.../controlPanelCommon/network/index.ts:640):
// `POST /middleware/v1/<type>/password` with body
// `{name, namespace, middleware, user, password}`.
//
// --password is OPTIONAL on the command line. If omitted, we prompt
// twice via golang.org/x/term.ReadPassword (no echo) and require both
// entries to match. This is the recommended path: shell history never
// captures the secret. Pass --password explicitly only when the
// caller is a script that already has the value securely.
//
// ConfirmDestructive prompts before issuing the POST, showing the
// type / namespace / name / user so the operator can spot a wrong
// instance before the rotation lands.
func NewSetCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		typeRaw   string
		name      string
		namespace string
		user      string
		password  string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "rotate the admin/user password on one middleware instance",
		Long: `Rotate the admin (or named user) password on one Olares middleware
instance.

REQUIRED flags:
  --type        middleware type (mongodb / postgres / redis / rabbitmq /
                  minio / nats / mysql / mariadb / elasticsearch).
  --name        instance name.
  --namespace   instance namespace.
  --user        target user (e.g. "admin" or any DB-internal username).

--password is OPTIONAL and SHOULD usually be omitted: when not
provided, you will be prompted twice (no echo) and the two entries
must match. Passing --password on the command line leaks the secret
into shell history — only do it from a wrapper script that already
controls the value securely.

The verb is wrapped in ConfirmDestructive (a wrong --name will
break the running instance until you can re-rotate). Pass --yes to
skip the prompt.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			t, err := normalizeType(typeRaw)
			if err != nil {
				return err
			}
			if name == "" || namespace == "" || user == "" {
				return fmt.Errorf("--name, --namespace, and --user are required")
			}
			pwd, err := resolvePassword(password)
			if err != nil {
				return err
			}
			return runSet(c.Context(), o, t, name, namespace, user, pwd, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&typeRaw, "type", "t", "", "middleware type (REQUIRED): "+strings.Join(supportedTypes, " | "))
	cmd.Flags().StringVar(&name, "name", "", "instance name (REQUIRED)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "instance namespace (REQUIRED)")
	cmd.Flags().StringVar(&user, "user", "", "target user (REQUIRED; e.g. admin)")
	cmd.Flags().StringVar(&password, "password", "", "new password; if omitted, prompted twice (no echo) — RECOMMENDED to omit")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	o.AddOutputFlags(cmd)
	return cmd
}

// normalizeType rejects unknown types so a typo doesn't reach the
// server (which would respond with a generic 404). The list is
// hardcoded against the SPA's MiddlewareType enum — if upstream
// adds a new type, this validator is the single place to update.
func normalizeType(s string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return "", fmt.Errorf("--type is required")
	}
	for _, k := range supportedTypes {
		if t == k {
			return t, nil
		}
	}
	return "", fmt.Errorf("unsupported --type %q (want one of: %s)", s, strings.Join(supportedTypes, ", "))
}

// resolvePassword returns either the explicit --password value or the
// double-prompted no-echo value. We require stdin to be a TTY for
// the prompt path; non-TTY without --password is a hard error
// rather than a silent prompt-then-EOF.
func resolvePassword(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("stdin is not a terminal — pass --password explicitly when running non-interactively")
	}
	fmt.Fprint(os.Stderr, "New password: ")
	first, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	fmt.Fprint(os.Stderr, "Confirm password: ")
	second, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("confirm password: %w", err)
	}
	if string(first) != string(second) {
		return "", fmt.Errorf("passwords do not match")
	}
	if len(first) == 0 {
		return "", fmt.Errorf("password must not be empty")
	}
	return string(first), nil
}

// passwordRequest mirrors MiddlewarePasswordParams in
// apps/.../middleware.ts. Field names are JSON-tagged exactly because
// the server validates on those names.
type passwordRequest struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Middleware string `json:"middleware"`
	User       string `json:"user"`
	Password   string `json:"password"`
}

// passwordResponse mirrors MiddlewarePasswordResponse — a custom
// envelope (not K8s shape) with a numeric code and human message.
type passwordResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// setResult is the JSON-mode shape emitted on success. We never echo
// the password back, even in JSON (security: --output json into a
// log file would be a leak).
type setResult struct {
	Operation  string `json:"operation"`
	Type       string `json:"type"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	User       string `json:"user"`
	ServerCode int    `json:"serverCode"`
}

func runSet(ctx context.Context, o *clusteropts.ClusterOptions, mwType, name, namespace, user, password string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Rotate password for %s instance %s/%s (user=%s)?", mwType, namespace, name, user),
		assumeYes); err != nil {
		return err
	}

	client, err := o.Prepare()
	if err != nil {
		return err
	}
	body := passwordRequest{
		Name:       name,
		Namespace:  namespace,
		Middleware: mwType,
		User:       user,
		Password:   password,
	}
	path := fmt.Sprintf("/middleware/v1/%s/password", url.PathEscape(mwType))
	var resp passwordResponse
	if err := client.DoJSON(ctx, "POST", path, body, &resp); err != nil {
		return fmt.Errorf("set %s password for %s/%s user=%s: %w", mwType, namespace, name, user, err)
	}
	if resp.Code != 0 && resp.Code != 200 {
		// Server returned a structured failure inside a 2xx HTTP
		// envelope. Surface code + message so the user sees the
		// actionable error.
		if resp.Message != "" {
			return fmt.Errorf("server returned code=%d: %s", resp.Code, resp.Message)
		}
		return fmt.Errorf("server returned code=%d", resp.Code)
	}

	result := setResult{
		Operation:  "password set",
		Type:       mwType,
		Namespace:  namespace,
		Name:       name,
		User:       user,
		ServerCode: resp.Code,
	}
	if o.IsJSON() {
		return o.PrintJSON(result)
	}
	if !o.Quiet {
		fmt.Fprintf(os.Stdout, "password rotated for %s instance %s/%s user=%s\n",
			mwType, namespace, name, user)
	}
	return nil
}
