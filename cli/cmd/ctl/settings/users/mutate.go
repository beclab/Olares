package users

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/me"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// NewCreateCommand implements `settings users create`.
func NewCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		role          string
		cpu           string
		memoryGB      int
		uiDefaults    bool
		displayName   string
		description   string
		email         string
		output        string
		watch         bool
		watchTimeout  time.Duration
		watchInterval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "create <username>",
		Short: "create an Olares user (Settings -> Users)",
		Long: `Create an Olares user (same flow as Settings → Users in Termipass).

Without --defaults you must set --role, --cpu, and --memory-gb explicitly.
Role may only be admin or normal (the SPA account dialog cannot create an owner).

With --defaults the CLI uses the same preset as the SPA form: role normal
(members), cpu 1, memory 4G (do not combine with those three flags).

The initial password is always auto-generated with the same rules as Termipass.

By default the CLI returns once user-service accepts the create request and
prints the Olares ID and one-time password immediately. Pass -w/--watch to
block until the user reaches Created and the Wizard URL is available — the
opt-in shape matches "olares-cli market <verb> --watch".

Human-oriented table mode (default) prints short progress hints to stderr; use
--output json if you want a quiet stderr for scripting.

Not the same as "olares-cli user create" (kube / cluster API).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "create user"); err != nil {
				return err
			}
			nameArg := strings.TrimSpace(args[0])
			if nameArg == "" {
				return fmt.Errorf("username is required")
			}

			var ownerRole string
			var cpuLimit string
			var memWire string
			if uiDefaults {
				if c.Flags().Changed("role") || c.Flags().Changed("cpu") || c.Flags().Changed("memory-gb") {
					return fmt.Errorf("do not combine --defaults with --role, --cpu, or --memory-gb")
				}
				ownerRole = "normal"
				cpuLimit = "1"
				memWire = "4G"
			} else {
				if !c.Flags().Changed("role") {
					return fmt.Errorf("--role is required (or use --defaults for normal / 1 cpu / 4G memory)")
				}
				ownerRole = strings.TrimSpace(role)
				if ownerRole == "" {
					return fmt.Errorf("--role must not be empty")
				}
				switch ownerRole {
				case "admin", "normal":
				default:
					return fmt.Errorf("role must be admin or normal (creating an owner account is not supported here)")
				}
				if !c.Flags().Changed("cpu") {
					return fmt.Errorf("--cpu is required (or use --defaults)")
				}
				cpuLimit = strings.TrimSpace(cpu)
				if cpuLimit == "" {
					return fmt.Errorf("--cpu must not be empty")
				}
				if !c.Flags().Changed("memory-gb") {
					return fmt.Errorf("--memory-gb is required (or use --defaults)")
				}
				if memoryGB <= 0 {
					return fmt.Errorf("--memory-gb must be a positive integer")
				}
				memWire = fmt.Sprintf("%dG", memoryGB)
			}

			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runCreate(ctx, f, createParams{
				name:        nameArg,
				ownerRole:   ownerRole,
				cpuLimit:    cpuLimit,
				memoryLimit: memWire,
				displayName: displayName,
				description: description,
				email:       email,
				format:      format,
				watch: watchOptions{
					watch:    watch,
					timeout:  watchTimeout,
					interval: watchInterval,
				},
			}), "create user")
		},
	}
	// Flag order: role, cpu, memory, defaults first in help; then operational flags; optional body fields hidden.
	cmd.Flags().StringVar(&role, "role", "", "account role: admin | normal (required unless --defaults)")
	cmd.Flags().StringVar(&cpu, "cpu", "", `CPU limit (e.g. "1") (required unless --defaults)`)
	cmd.Flags().IntVar(&memoryGB, "memory-gb", -1, "memory in GB, sent as <n>G (required unless --defaults)")
	cmd.Flags().BoolVar(&uiDefaults, "defaults", false, "use SPA form preset: role normal, cpu 1, memory 4G (mutually exclusive with --role, --cpu, --memory-gb)")

	cmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	_ = cmd.Flags().MarkHidden("display-name")
	cmd.Flags().StringVar(&description, "description", "", "description")
	_ = cmd.Flags().MarkHidden("description")
	cmd.Flags().StringVar(&email, "email", "", "email")
	_ = cmd.Flags().MarkHidden("email")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false,
		"block until the user reaches Created (or Failed) before exiting; same opt-in shape as 'olares-cli market <verb> --watch'")
	cmd.Flags().DurationVar(&watchTimeout, "watch-timeout", 15*time.Minute,
		"maximum total time to wait when --watch is set (default 15m, matches market)")
	cmd.Flags().DurationVar(&watchInterval, "watch-interval", 2*time.Second,
		"polling interval for /status when --watch is set (default 2s, matches market)")
	addOutputFlag(cmd, &output)

	cmd.Flags().SortFlags = false
	return cmd
}

// watchOptions is the per-command knob bundle for the opt-in `--watch`
// polling mode. Mirrors cli/cmd/ctl/market/options.go addWatchFlags so
// both surfaces present the same flag triple to operators.
type watchOptions struct {
	watch    bool
	timeout  time.Duration
	interval time.Duration
}

type createParams struct {
	name, ownerRole, cpuLimit, memoryLimit string
	displayName, description, email        string
	format                                 Format
	watch                                  watchOptions
}

// accountModifyStatus mirrors user-service/account_status proxied body
// (framework/app-service userStatus handler).
type accountModifyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Address struct {
		Wizard string `json:"wizard"`
	} `json:"address"`
}

func runCreate(ctx context.Context, f *cmdutil.Factory, p createParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var probe olaresInfoProbe
	if err := decodeObjectResult(ctx, pc.Doer, "/api/olares-info", &probe); err != nil {
		return fmt.Errorf("read Olares/cluster info needed for hashing and identity check: %w", err)
	}

	humanProgress := p.format != FormatJSON
	if humanProgress {
		fmt.Fprintf(os.Stderr, "[create user] verifying new account identity …\n")
	}

	if err := precheckNewUserOlaresIDDID(ctx, pc, probe, p.name); err != nil {
		return err
	}

	if humanProgress {
		fmt.Fprintf(os.Stderr, "[create user] sending create request …\n")
	}

	rawPWD, err := generatePasswordSPA()
	if err != nil {
		return err
	}
	wirePWD := me.SaltedPassword(rawPWD, probe.OsVersion)

	body := map[string]string{
		"name":         p.name,
		"owner_role":   p.ownerRole,
		"password":     wirePWD,
		"cpu_limit":    p.cpuLimit,
		"memory_limit": p.memoryLimit,
	}
	if s := strings.TrimSpace(p.displayName); s != "" {
		body["display_name"] = s
	}
	if s := strings.TrimSpace(p.description); s != "" {
		body["description"] = s
	}
	if s := strings.TrimSpace(p.email); s != "" {
		body["email"] = s
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := doMutateUsersAPI(ctx, pc.Doer, "POST", "/api/users", body, &resp); err != nil {
		return err
	}
	createdName := p.name
	if strings.TrimSpace(resp.Name) != "" {
		createdName = strings.TrimSpace(resp.Name)
	}

	if humanProgress && p.watch.watch {
		q := formatDurationBrief(p.watch.interval)
		tmax := formatDurationBrief(p.watch.timeout)
		fmt.Fprintf(os.Stderr, "[create user] create accepted for %q: waiting for provisioning to finish (check every %s, timeout %s) …\n",
			createdName, q, tmax)
	}

	var (
		wizardHost  string
		finalStatus string
	)
	if p.watch.watch {
		st, err := waitForUserState(ctx, pc.Doer, userWatchOptions{
			Timeout:  p.watch.timeout,
			Interval: p.watch.interval,
			Progress: humanProgress,
		}, newUserWatchTarget(userWatchCreate, createdName))
		if err != nil {
			return err
		}
		if humanProgress {
			fmt.Fprintf(os.Stderr, "[create user] provisioning finished; user is ready.\n")
		}
		wizardHost = strings.TrimSpace(st.Address.Wizard)
		finalStatus = strings.TrimSpace(st.Status)
	}

	wizardURL := ""
	if wizardHost != "" {
		wizardURL = "https://" + strings.TrimPrefix(strings.TrimPrefix(wizardHost, "https://"), "http://")
	}

	switch p.format {
	case FormatJSON:
		out := map[string]string{
			"name":              createdName,
			"original_password": rawPWD,
		}
		if p.watch.watch {
			out["status"] = "Created"
			if finalStatus != "" {
				out["final_status"] = finalStatus
			}
			if wizardURL != "" {
				out["wizard_url"] = wizardURL
			}
		} else {
			out["status"] = "Accepted"
		}
		return printJSON(os.Stdout, out)
	default:
		return printCreateSuccessTTY(os.Stdout, createdName, rawPWD, wizardURL, p.watch.watch)
	}
}

// httpStatusFromErrHint detects HTTP status digits in backend errors (see whoami.formatBackendErr).
func httpStatusFromErrHint(err error, code int) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", code))
}

// formatDurationBrief prints poll/timeout hints without scientific notation.
func formatDurationBrief(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	rd := d.Round(time.Second)
	s := rd.String()
	s = strings.ReplaceAll(s, "h0m0s", "h")
	s = strings.ReplaceAll(s, "m0s", "m")
	return s
}

// printCreateSuccessTTY renders the human-readable summary after a create
// request. `watched` reflects whether the caller blocked on /status until
// Created (-w/--watch); when false the Wizard URL was deliberately not
// fetched and the user must re-query later.
func printCreateSuccessTTY(w io.Writer, username, rawPassword, wizardURL string, watched bool) error {
	buf := strings.Builder{}
	if watched {
		buf.WriteString("User had been created.\n\n")
	} else {
		buf.WriteString("Create request accepted; provisioning continues asynchronously.\n\n")
	}
	fmt.Fprintf(&buf, "Olares ID:          %s\n", username)
	fmt.Fprintf(&buf, "Original password:  %s\n", rawPassword)
	if wizardURL != "" {
		fmt.Fprintf(&buf, "Wizard URL:         %s\n", wizardURL)
	} else if !watched {
		buf.WriteString("Wizard URL:         (not fetched; pass --watch to wait until provisioning completes, then re-run with --watch or use \"olares-cli settings users get <name>\")\n")
	} else {
		buf.WriteString("Wizard URL:         (empty — check status later)\n")
	}
	buf.WriteString("\nSave this information; it is shown once.\n")
	_, err := fmt.Fprint(w, buf.String())
	return err
}

// NewDeleteCommand implements `settings users delete`.
func NewDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		skipConfirm   bool
		output        string
		watch         bool
		watchTimeout  time.Duration
		watchInterval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "delete <username>",
		Short: "delete an Olares user (Settings -> Users)",
		Long: `Delete an Olares user (same flow as Settings → Users in Termipass).

By default type the whole word yes when prompted (like ssh). Use --yes only
for scripting (skip confirmation).

By default the CLI returns once user-service accepts the DELETE request (the
row may still appear briefly in "users list" while controllers tear it down).
Pass -w/--watch to block until removal is fully reported as finished — same
opt-in shape as "olares-cli market <verb> --watch".

Use --output json when you want a quiet stderr stream for scripting.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "delete user"); err != nil {
				return err
			}
			u := strings.TrimSpace(args[0])
			if u == "" {
				return fmt.Errorf("username is required")
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runDelete(ctx, f, deleteParams{
				username:    u,
				skipConfirm: skipConfirm,
				format:      format,
				watch: watchOptions{
					watch:    watch,
					timeout:  watchTimeout,
					interval: watchInterval,
				},
			}), "delete user")
		},
	}
	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "skip interactive confirmation (dangerous)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false,
		"block until the user reaches Deleted before exiting; same opt-in shape as 'olares-cli market <verb> --watch'")
	cmd.Flags().DurationVar(&watchTimeout, "watch-timeout", 15*time.Minute,
		"maximum total time to wait when --watch is set (default 15m, matches market)")
	cmd.Flags().DurationVar(&watchInterval, "watch-interval", 2*time.Second,
		"polling interval for /status when --watch is set (default 2s, matches market)")
	addOutputFlag(cmd, &output)
	return cmd
}

type deleteParams struct {
	username    string
	skipConfirm bool
	format      Format
	watch       watchOptions
}

func userIsOwner(info *userInfo) bool {
	if info == nil {
		return false
	}
	for _, r := range info.Roles {
		if strings.TrimSpace(r) == "owner" {
			return true
		}
	}
	return false
}

func fetchUserForDelete(ctx context.Context, d Doer, username string) (*userInfo, error) {
	path := "/api/users/" + url.PathEscape(username)
	var info userInfo
	if err := decodeObjectResult(ctx, d, path, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func validateUserDeletable(username string, info *userInfo) error {
	if userIsOwner(info) {
		return fmt.Errorf("cannot delete user '%s' with role '%s' ", username, "owner")
	}
	if strings.TrimSpace(info.State) == "Deleting" {
		return fmt.Errorf("user %q is already being deleted", username)
	}
	return nil
}

func runDelete(ctx context.Context, f *cmdutil.Factory, p deleteParams) error {
	if ctx == nil {
		ctx = context.Background()
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	info, err := fetchUserForDelete(ctx, pc.Doer, p.username)
	if err != nil {
		return err
	}
	if err := validateUserDeletable(p.username, info); err != nil {
		return err
	}

	if !p.skipConfirm {
		fmt.Fprintf(os.Stderr, "This will permanently delete Olares user %q.\n"+
			"Type 'yes' to continue: ", p.username)
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) != "yes" {
			return fmt.Errorf("aborting delete: confirmation was not yes")
		}
	}

	path := "/api/users/" + url.PathEscape(p.username)
	var resp struct {
		Name string `json:"name"`
	}
	if err := doMutateUsersAPI(ctx, pc.Doer, "DELETE", path, nil, &resp); err != nil {
		return err
	}
	deletedName := p.username
	if strings.TrimSpace(resp.Name) != "" {
		deletedName = strings.TrimSpace(resp.Name)
	}

	humanProgress := p.format != FormatJSON
	if humanProgress && !p.watch.watch {
		fmt.Fprintf(os.Stderr, "[delete user] delete accepted for %q; not waiting for cleanup (pass --watch to block until Deleted).\n",
			deletedName)
	}
	if humanProgress && p.watch.watch {
		fmt.Fprintf(os.Stderr, "[delete user] delete accepted for %q: waiting until removal finishes (check every %s, timeout %s) …\n",
			deletedName, formatDurationBrief(p.watch.interval), formatDurationBrief(p.watch.timeout))
	}
	var finalStatus string
	if p.watch.watch {
		st, err := waitForUserState(ctx, pc.Doer, userWatchOptions{
			Timeout:  p.watch.timeout,
			Interval: p.watch.interval,
			Progress: humanProgress,
		}, newUserWatchTarget(userWatchDelete, deletedName))
		if err != nil {
			return err
		}
		if humanProgress {
			fmt.Fprintf(os.Stderr, "[delete user] removal finished.\n")
		}
		if st != nil {
			finalStatus = strings.TrimSpace(st.Status)
		}
	}

	switch p.format {
	case FormatJSON:
		out := map[string]string{
			"name": deletedName,
		}
		if p.watch.watch {
			out["status"] = "Deleted"
			if finalStatus != "" {
				out["final_status"] = finalStatus
			}
		} else {
			out["status"] = "Accepted"
		}
		return printJSON(os.Stdout, out)
	default:
		nameOut := deletedName
		if nameOut == "" {
			_, err := fmt.Fprintln(os.Stdout, "delete request accepted")
			return err
		}
		if p.watch.watch {
			_, err := fmt.Fprintf(os.Stdout, "deleted user %q\n", nameOut)
			return err
		}
		_, err := fmt.Fprintf(os.Stdout, "delete request accepted for %q (still removing asynchronously; pass --watch next time to block until Deleted)\n", nameOut)
		return err
	}
}
