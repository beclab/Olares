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
		noWait        bool
		provisionTO   time.Duration
		provisionPoll time.Duration
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

The CLI verifies the new account identity first, then waits until provisioning
finishes (unless --no-wait), then prints Olares ID, one-time password, and Wizard URL.

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
				name:            nameArg,
				ownerRole:       ownerRole,
				cpuLimit:        cpuLimit,
				memoryLimit:     memWire,
				displayName:     displayName,
				description:     description,
				email:           email,
				format:          format,
				noWaitProvision: noWait,
				provisionTO:     provisionTO,
				provisionPoll:   provisionPoll,
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
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "do not poll until Created; print password immediately without wizard URL")
	cmd.Flags().DurationVar(&provisionTO, "provision-timeout", 30*time.Minute, "max time to wait for provisioning to finish")
	cmd.Flags().DurationVar(&provisionPoll, "provision-poll", 4*time.Second, "how often to check provisioning progress while waiting")
	addOutputFlag(cmd, &output)

	cmd.Flags().SortFlags = false
	return cmd
}

type createParams struct {
	name, ownerRole, cpuLimit, memoryLimit string
	displayName, description, email        string
	format                                 Format
	noWaitProvision                        bool
	provisionTO, provisionPoll             time.Duration
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

	if humanProgress && !p.noWaitProvision {
		q := formatDurationBrief(p.provisionPoll)
		tmax := formatDurationBrief(p.provisionTO)
		fmt.Fprintf(os.Stderr, "[create user] create accepted for %q: waiting for provisioning to finish (check every %s, timeout %s) …\n",
			createdName, q, tmax)
	}

	var wizardHost string
	if !p.noWaitProvision {
		st, err := waitForUserCreated(ctx, pc.Doer, createdName, humanProgress,
			p.provisionPoll, p.provisionTO)
		if err != nil {
			return err
		}
		if humanProgress {
			fmt.Fprintf(os.Stderr, "[create user] provisioning finished; user is ready.\n")
		}
		wizardHost = strings.TrimSpace(st.Address.Wizard)
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
			"status":            "Created",
		}
		if wizardURL != "" {
			out["wizard_url"] = wizardURL
		} else if p.noWaitProvision {
			out["status"] = "Accepted"
			out["note"] = "provisioning was not waited on; use \"settings users get <name>\" or run without --no-wait to wait until the new user is fully ready"
		}
		return printJSON(os.Stdout, out)
	default:
		return printCreateSuccessTTY(os.Stdout, createdName, rawPWD, wizardURL, p.noWaitProvision)
	}
}

func waitForUserCreated(ctx context.Context, d Doer, username string, progress bool, poll, timeout time.Duration) (*accountModifyStatus, error) {
	if poll <= 0 {
		poll = 4 * time.Second
	}
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	path := "/api/users/" + url.PathEscape(username) + "/status"
	deadline := time.Now().Add(timeout)
	lastStatus := ""
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var st accountModifyStatus
		if err := decodeObjectResult(ctx, d, path, &st); err != nil {
			return nil, err
		}
		lastStatus = strings.TrimSpace(st.Status)
		switch lastStatus {
		case "Created":
			return &st, nil
		case "Failed":
			msg := strings.TrimSpace(st.Message)
			if msg == "" {
				msg = "upstream reported Failed with no message"
			}
			return nil, fmt.Errorf("user provisioning failed: %s", msg)
		case "Deleted":
			return nil, fmt.Errorf("user %q disappeared while provisioning (status Deleted)", username)
		default:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for user %q to reach Created (last status=%q)", username, lastStatus)
			}
			if progress {
				label := lastStatus
				if label == "" {
					label = "…"
				}
				fmt.Fprintf(os.Stderr, "[create user] provisioning in progress for %q (lifecycle state reported by server: %s); next check in %s …\n",
					username, label, formatDurationBrief(poll))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(poll):
			}
		}
	}
}

func waitForUserDeleted(ctx context.Context, d Doer, username string, progress bool, poll, timeout time.Duration) (*accountModifyStatus, error) {
	if poll <= 0 {
		poll = 4 * time.Second
	}
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	path := "/api/users/" + url.PathEscape(username) + "/status"
	deadline := time.Now().Add(timeout)
	lastStatus := ""
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var st accountModifyStatus
		err := decodeObjectResult(ctx, d, path, &st)
		if err != nil {
			if httpStatusFromErrHint(err, 404) {
				st.Name = username
				st.Status = "Deleted"
				return &st, nil
			}
			if httpStatusFromErrHint(err, 401) || httpStatusFromErrHint(err, 403) {
				return nil, err
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf(
					"timeout waiting for user %q to reach Deleted while polling status (last known status=%q): %w",
					username, lastStatus, err)
			}
			if progress {
				fmt.Fprintf(os.Stderr,
					"[delete user] transient status poll error (%v); retry in %s …\n",
					err, formatDurationBrief(poll))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(poll):
			}
			continue
		}

		lastStatus = strings.TrimSpace(st.Status)
		switch lastStatus {
		case "Deleted":
			return &st, nil
		default:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for user %q to reach Deleted (last status=%q)", username, lastStatus)
			}
			if progress {
				label := lastStatus
				if label == "" {
					label = "…"
				}
				fmt.Fprintf(os.Stderr, "[delete user] removal in progress for %q (lifecycle state reported by server: %s); next check in %s …\n",
					username, label, formatDurationBrief(poll))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(poll):
			}
		}
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

func printCreateSuccessTTY(w io.Writer, username, rawPassword, wizardURL string, noWait bool) error {
	buf := strings.Builder{}
	buf.WriteString("User had been created.\n\n")
	fmt.Fprintf(&buf, "Olares ID:          %s\n", username)
	fmt.Fprintf(&buf, "Original password:  %s\n", rawPassword)
	if wizardURL != "" {
		fmt.Fprintf(&buf, "Wizard URL:         %s\n", wizardURL)
	} else if noWait {
		buf.WriteString("Wizard URL:         (not fetched; --no-wait — run \"settings users get <name>\" after provisioning finishes)\n")
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
		noWait        bool
		deletePoll    time.Duration
		deleteTimeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "delete <username>",
		Short: "delete an Olares user (Settings -> Users)",
		Long: `Delete an Olares user (same flow as Settings → Users in Termipass).

By default type the whole word yes when prompted (like ssh). Use --yes only
for scripting (skip confirmation).

By default the CLI waits until removal is fully reported as finished (same pacing
as Termipass); use --no-wait to exit as soon as the stop request is accepted
(the account may still disappear from the list asynchronously).

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
				noWaitPoll:  noWait,
				poll:        deletePoll,
				timeout:     deleteTimeout,
			}), "delete user")
		},
	}
	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "skip interactive confirmation (dangerous)")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "after delete is accepted, do not wait for removal to finish")
	cmd.Flags().DurationVar(&deletePoll, "delete-poll", 4*time.Second, "how often to check whether removal has finished while waiting")
	cmd.Flags().DurationVar(&deleteTimeout, "delete-timeout", 30*time.Minute, "max time to wait for removal to finish after delete is accepted")
	addOutputFlag(cmd, &output)
	return cmd
}

type deleteParams struct {
	username      string
	skipConfirm   bool
	format        Format
	noWaitPoll    bool
	poll, timeout time.Duration
}

func runDelete(ctx context.Context, f *cmdutil.Factory, p deleteParams) error {
	if ctx == nil {
		ctx = context.Background()
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

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
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
	if humanProgress && p.noWaitPoll {
		fmt.Fprintf(os.Stderr, "[delete user] not waiting for cleanup to finish (--no-wait); \"%s\" may still appear in the list briefly.\n",
			deletedName)
	}
	if humanProgress && !p.noWaitPoll {
		fmt.Fprintf(os.Stderr, "[delete user] delete accepted for %q: waiting until removal finishes (check every %s, timeout %s) …\n",
			deletedName, formatDurationBrief(p.poll), formatDurationBrief(p.timeout))
	}
	if !p.noWaitPoll {
		if _, err := waitForUserDeleted(ctx, pc.Doer, deletedName, humanProgress, p.poll, p.timeout); err != nil {
			return err
		}
		if humanProgress {
			fmt.Fprintf(os.Stderr, "[delete user] removal finished.\n")
		}
	}

	switch p.format {
	case FormatJSON:
		out := map[string]string{
			"name": deletedName,
		}
		if p.noWaitPoll {
			out["status"] = "Deleting"
			out["note"] = "delete was accepted only; removal may still finish in the background — re-run this command without --no-wait to wait until removal completes"
		} else {
			out["status"] = "Deleted"
		}
		return printJSON(os.Stdout, out)
	default:
		nameOut := deletedName
		if nameOut != "" {
			_, err := fmt.Fprintf(os.Stdout, "deleted user %q\n", nameOut)
			return err
		}
		_, err := fmt.Fprintln(os.Stdout, "user deleted")
		return err
	}
}
