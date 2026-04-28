package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps policy get|set <app> <entrance>`.
//
// Per-entrance authorization policy (the "Two-factor" panel in the SPA
// at apps/.../ApplicationEntrancePage.vue). Mirrors the SPA's
// stores/settings/application.ts {getPolicy, set_appFa2} pair.
//
// Wire shape (BFL-proxied; no user-service controller):
//
//   GET  /api/applications/<app>/<entrance>/setup/policy
//        → { default_policy, one_time, valid_duration, sub_policies[] }
//
//   POST /api/applications/<app>/<entrance>/setup/policy
//        body: { default_policy, one_time, valid_duration, sub_policies }
//        — replaces the whole policy. sub_policies is null when there
//          are zero entries (the SPA explicitly converts [] → null
//          before posting; we mirror that).
//
// Policy values (FACTOR_MODEL enum from constant/index.ts):
//
//   "system"      — inherit the system default policy
//   "one_factor"  — single-factor (password) auth
//   "two_factor"  — two-factor (password + TOTP) auth
//   "public"      — no auth required
//
// Sub-policy entries (one per URI) override the default policy for
// specific paths under the entrance:
//
//   { one_time, policy, uri, valid_duration }
//
// The CLI accepts sub_policies via:
//
//   --sub-policy "<spec>"   repeatable; comma-separated key=value pairs.
//                            Required: uri=<path> + policy=<one of above>.
//                            Optional: one_time=true|false (default false),
//                                       valid_duration=<seconds> (default 0).
//   --sub-policies-file <path>
//                            JSON array of {one_time, policy, uri,
//                            valid_duration} entries. Mutually exclusive
//                            with --sub-policy.
//   --clear-sub-policies     post sub_policies as null (matches the
//                            SPA's "no entries" path).
//
// Default-policy / one-time / valid-duration use the read-modify-write
// pattern: any field not explicitly passed survives untouched, mirroring
// what the SPA submits when the user only changes one knob in the
// dialog.
//
// Role: per-app config writes; the SPA gates on isAdmin. We rely on
// server-side preflight (a normal user gets a 403 with the usual hint).

// EntrancePolicy mirrors the SPA's EntrancePolicy interface
// (apps/.../constant/index.ts:187-192) — a single sub-policy row.
type EntrancePolicy struct {
	OneTime       bool   `json:"one_time"`
	Policy        string `json:"policy"`
	URI           string `json:"uri"`
	ValidDuration int    `json:"valid_duration"`
}

// SetupPolicy mirrors the GET /setup/policy response. sub_policies can
// be either nil OR an array; the SPA treats nil + [] interchangeably.
type SetupPolicy struct {
	DefaultPolicy string           `json:"default_policy"`
	OneTime       bool             `json:"one_time"`
	ValidDuration int              `json:"valid_duration"`
	SubPolicies   []EntrancePolicy `json:"sub_policies"`
}

// setupPolicyBody is the POST body. SubPolicies is *[]... so we can
// distinguish nil (sent as JSON null, matching SPA's "no entries" path)
// from an empty slice that some servers might reject differently.
type setupPolicyBody struct {
	DefaultPolicy string            `json:"default_policy"`
	OneTime       bool              `json:"one_time"`
	ValidDuration int               `json:"valid_duration"`
	SubPolicies   *[]EntrancePolicy `json:"sub_policies"`
}

// validPolicies covers the four FACTOR_MODEL values the SPA accepts.
// We validate client-side so users get a quicker, friendlier error than
// the upstream's generic 400.
var validPolicies = map[string]struct{}{
	"system":     {},
	"one_factor": {},
	"two_factor": {},
	"public":     {},
}

// NewPolicyCommand returns the `settings apps policy` parent.
func NewPolicyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "per-entrance authorization policy (Settings -> App -> Entrance -> Policy)",
		Long: `Manage the per-entrance authorization policy — the default factor
mode (one_factor / two_factor / system / public) plus optional sub
policies that override per URI.

Subcommands:
  list <app>                                          list every entrance's policy
  get  <app> <entrance>                               show current policy
  set  <app> <entrance> [--default-policy MODE]
                        [--one-time true|false]
                        [--valid-duration SECONDS]
                        [--sub-policy "uri=...,policy=..."] (repeatable)
                        [--sub-policies-file PATH]
                        [--clear-sub-policies]
                                                      replace the policy (RMW)

If you only know the app name and not its entrances yet, run
"apps policy list <app>" (or "apps entrances list <app>") first to
discover them.

--default-policy values: system | one_factor | two_factor | public

Set semantics: unspecified flags survive (read-modify-write).
Sub-policy entries are REPLACED in full whenever any sub-policy flag
is passed; pass --clear-sub-policies to drop them.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPolicyListCommand(f))
	cmd.AddCommand(newPolicyGetCommand(f))
	cmd.AddCommand(newPolicySetCommand(f))
	return cmd
}

// newPolicyListCommand registers `apps policy list <app>`. Same shape /
// motivation as `apps domain list`: fan out one GET /setup/policy per
// entrance so users without a known entrance name can land on
// per-entrance results in one command. Closes the KI-17 "policy get
// <app>" complaint by introducing a "list <app>" sibling rather than
// loosening the strict 2-arg signature of "get".
func newPolicyListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <app>",
		Short: "list auth policy for every entrance of an app",
		Long: `List the per-entrance auth policies (default mode + sub-policies) for
every entrance the app exposes. Internally fans out one "policy get"
per entrance and renders them in series.

Pass --output json for an array of {entrance, policy} objects.

Examples:
  olares-cli settings apps policy list files
  olares-cli settings apps policy list files -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPolicyList(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// policyListEntry pairs an entrance name with its SetupPolicy for the
// JSON output of `apps policy list`.
type policyListEntry struct {
	Entrance string      `json:"entrance"`
	Policy   SetupPolicy `json:"policy"`
}

func runPolicyList(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	entrancePath := "/api/applications/" + url.PathEscape(app) + "/entrances"
	var env entrancesEnvelope
	if err := doGetEnvelope(ctx, pc.doer, entrancePath, &env); err != nil {
		return err
	}
	entries := make([]policyListEntry, 0, len(env.Items))
	for _, e := range env.Items {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		sp, err := getPolicy(ctx, pc.doer, app, name)
		if err != nil {
			return fmt.Errorf("entrance %q: %w", name, err)
		}
		entries = append(entries, policyListEntry{Entrance: name, Policy: sp})
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, entries)
	}
	return renderPolicyList(os.Stdout, app, entries)
}

func renderPolicyList(w io.Writer, app string, entries []policyListEntry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintf(w, "no entrances for app %q\n", app)
		return err
	}
	for i, e := range entries {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := renderPolicy(w, app, e.Entrance, e.Policy); err != nil {
			return err
		}
	}
	return nil
}

func newPolicyGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <app> <entrance>",
		Short: "show current per-entrance auth policy",
		Long: `Show the default policy mode + sub-policies for the entrance.

Pass --output json for the raw SetupPolicy struct including the full
sub-policies vector.

Examples:
  olares-cli settings apps policy get files file
  olares-cli settings apps policy get files file -o json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runPolicyGet(c.Context(), f, args[0], args[1], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runPolicyGet(ctx context.Context, f *cmdutil.Factory, app, entrance, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app, entrance = strings.TrimSpace(app), strings.TrimSpace(entrance)
	if app == "" || entrance == "" {
		return fmt.Errorf("both <app> and <entrance> are required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	current, err := getPolicy(ctx, pc.doer, app, entrance)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, current)
	default:
		return renderPolicy(os.Stdout, app, entrance, current)
	}
}

func getPolicy(ctx context.Context, d Doer, app, entrance string) (SetupPolicy, error) {
	var sp SetupPolicy
	if err := doGetEnvelope(ctx, d, policyPath(app, entrance), &sp); err != nil {
		return SetupPolicy{}, err
	}
	return sp, nil
}

func renderPolicy(w io.Writer, app, entrance string, sp SetupPolicy) error {
	if _, err := fmt.Fprintf(w, "App:                 %s\n", app); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Entrance:            %s\n", entrance); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Default policy:      %s\n", nonEmpty(sp.DefaultPolicy)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "One-time:            %s\n", boolStr(sp.OneTime)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Valid duration (s):  %d\n", sp.ValidDuration); err != nil {
		return err
	}
	if len(sp.SubPolicies) == 0 {
		_, err := fmt.Fprintln(w, "Sub-policies:        (none)")
		return err
	}
	fmt.Fprintln(w, "Sub-policies:")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "  URI\tPOLICY\tONE-TIME\tVALID DURATION"); err != nil {
		return err
	}
	for _, e := range sp.SubPolicies {
		if _, err := fmt.Fprintf(tw, "  %s\t%s\t%s\t%d\n",
			nonEmpty(e.URI),
			nonEmpty(e.Policy),
			boolStr(e.OneTime),
			e.ValidDuration,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func newPolicySetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		defaultPolicy     string
		oneTime           bool
		validDuration     int
		subPolicySpecs    []string
		subPoliciesFile   string
		clearSubPolicies  bool
	)
	cmd := &cobra.Command{
		Use:   "set <app> <entrance>",
		Short: "replace per-entrance auth policy (read-modify-write)",
		Long: `Replace the per-entrance authorization policy. Unspecified flags
survive untouched (RMW); sub-policy flags REPLACE the existing
sub-policies vector when any of them is passed.

--default-policy values: system | one_factor | two_factor | public

Sub-policy specs (--sub-policy, repeatable) accept comma-separated
key=value pairs:

   uri=<path>            required, the URI prefix this rule overrides
   policy=<mode>         required, one of system|one_factor|two_factor|public
   one_time=true|false   optional, default false
   valid_duration=<s>    optional, default 0

Use --sub-policies-file <path> to load a JSON array directly. Mutually
exclusive with --sub-policy.

Examples:
  # Switch the entrance to two-factor by default, keep sub-policies
  olares-cli settings apps policy set files file --default-policy two_factor

  # Add two URI overrides via repeated --sub-policy
  olares-cli settings apps policy set files file \
      --sub-policy "uri=/admin,policy=two_factor,one_time=true,valid_duration=300" \
      --sub-policy "uri=/healthz,policy=public"

  # Drop all sub-policies, keep default
  olares-cli settings apps policy set files file --clear-sub-policies

  # Replace via JSON file
  olares-cli settings apps policy set files file --sub-policies-file ./policies.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runPolicySet(c.Context(), f, args[0], args[1], policySetFlags{
				defaultPolicy:    defaultPolicy,
				defaultPolicySet: c.Flags().Changed("default-policy"),
				oneTime:          oneTime,
				oneTimeSet:       c.Flags().Changed("one-time"),
				validDuration:    validDuration,
				validDurationSet: c.Flags().Changed("valid-duration"),
				subPolicySpecs:   subPolicySpecs,
				subPolicySet:     c.Flags().Changed("sub-policy"),
				subPoliciesFile:  subPoliciesFile,
				clearSubPolicies: clearSubPolicies,
			})
		},
	}
	cmd.Flags().StringVar(&defaultPolicy, "default-policy", "", "default factor mode (system|one_factor|two_factor|public)")
	cmd.Flags().BoolVar(&oneTime, "one-time", false, "require one-time auth across all sessions")
	cmd.Flags().IntVar(&validDuration, "valid-duration", 0, "auth token validity in seconds")
	cmd.Flags().StringArrayVar(&subPolicySpecs, "sub-policy", nil, "sub-policy spec (repeatable; comma-separated key=value pairs)")
	cmd.Flags().StringVar(&subPoliciesFile, "sub-policies-file", "", "load sub-policies from a JSON array file")
	cmd.Flags().BoolVar(&clearSubPolicies, "clear-sub-policies", false, "drop all sub-policies (post sub_policies: null)")
	return cmd
}

type policySetFlags struct {
	defaultPolicy    string
	defaultPolicySet bool
	oneTime          bool
	oneTimeSet       bool
	validDuration    int
	validDurationSet bool
	subPolicySpecs   []string
	subPolicySet     bool
	subPoliciesFile  string
	clearSubPolicies bool
}

func runPolicySet(ctx context.Context, f *cmdutil.Factory, app, entrance string, flags policySetFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	return runPolicySetWithDoer(ctx, pc.doer, app, entrance, flags)
}

// runPolicySetWithDoer is the wire-level core of `apps policy set`.
// Split out so unit tests can drive it through a fakeDoer without
// faking the whole cmdutil.Factory + credential plumbing. The flag
// validation and RMW merge live here; runPolicySet is now a thin
// wrapper that resolves the profile and forwards.
func runPolicySetWithDoer(ctx context.Context, d Doer, app, entrance string, flags policySetFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app, entrance = strings.TrimSpace(app), strings.TrimSpace(entrance)
	if app == "" || entrance == "" {
		return fmt.Errorf("both <app> and <entrance> are required")
	}
	if flags.defaultPolicySet {
		v := strings.TrimSpace(flags.defaultPolicy)
		if _, ok := validPolicies[v]; !ok {
			return fmt.Errorf("--default-policy %q is not one of system|one_factor|two_factor|public", flags.defaultPolicy)
		}
		flags.defaultPolicy = v
	}
	if flags.subPolicySet && flags.subPoliciesFile != "" {
		return fmt.Errorf("--sub-policy and --sub-policies-file are mutually exclusive")
	}
	if flags.clearSubPolicies && (flags.subPolicySet || flags.subPoliciesFile != "") {
		return fmt.Errorf("--clear-sub-policies cannot be combined with --sub-policy / --sub-policies-file")
	}
	if !flags.defaultPolicySet && !flags.oneTimeSet && !flags.validDurationSet &&
		!flags.subPolicySet && flags.subPoliciesFile == "" && !flags.clearSubPolicies {
		return fmt.Errorf("nothing to do — pass at least one of --default-policy / --one-time / --valid-duration / --sub-policy / --sub-policies-file / --clear-sub-policies")
	}

	current, err := getPolicy(ctx, d, app, entrance)
	if err != nil {
		return err
	}

	body := setupPolicyBody{
		DefaultPolicy: current.DefaultPolicy,
		OneTime:       current.OneTime,
		ValidDuration: current.ValidDuration,
	}
	// Preserve current sub-policies by default; the various
	// sub-policy flags reassign this pointer below if engaged.
	preserved := append([]EntrancePolicy(nil), current.SubPolicies...)
	body.SubPolicies = subPoliciesPtr(preserved)

	if flags.defaultPolicySet {
		body.DefaultPolicy = flags.defaultPolicy
	}
	if flags.oneTimeSet {
		body.OneTime = flags.oneTime
	}
	if flags.validDurationSet {
		body.ValidDuration = flags.validDuration
	}
	switch {
	case flags.clearSubPolicies:
		// SPA explicitly sends null when there are zero entries.
		body.SubPolicies = nil
	case flags.subPoliciesFile != "":
		parsed, parseErr := loadSubPoliciesFromFile(flags.subPoliciesFile)
		if parseErr != nil {
			return parseErr
		}
		body.SubPolicies = subPoliciesPtr(parsed)
	case flags.subPolicySet:
		parsed, parseErr := parseSubPolicySpecs(flags.subPolicySpecs)
		if parseErr != nil {
			return parseErr
		}
		body.SubPolicies = subPoliciesPtr(parsed)
	}

	if err := doMutateEnvelope(ctx, d, "POST", policyPath(app, entrance), body, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "updated auth policy for %s/%s (default=%s one_time=%t valid_duration=%d sub_policies=%s)\n",
		app, entrance, body.DefaultPolicy, body.OneTime, body.ValidDuration, summarizeSubPolicies(body.SubPolicies))
	return nil
}

// subPoliciesPtr converts a slice into the *[]EntrancePolicy our wire
// shape uses. An empty (but non-nil) slice still marshals as `[]` —
// upstream tolerates that — while nil marshals as `null`.
func subPoliciesPtr(in []EntrancePolicy) *[]EntrancePolicy {
	if in == nil {
		empty := []EntrancePolicy{}
		return &empty
	}
	return &in
}

// parseSubPolicySpecs accepts the user-friendly comma-separated form
// the --sub-policy flag takes. We tokenize on commas (plain — no
// support for embedded commas in URIs because that would require a
// proper escape syntax and the SPA's UI never produces such URIs
// anyway).
func parseSubPolicySpecs(specs []string) ([]EntrancePolicy, error) {
	if len(specs) == 0 {
		return []EntrancePolicy{}, nil
	}
	out := make([]EntrancePolicy, 0, len(specs))
	for i, spec := range specs {
		entry, err := parseSubPolicySpec(spec)
		if err != nil {
			return nil, fmt.Errorf("--sub-policy[%d] %q: %w", i, spec, err)
		}
		out = append(out, entry)
	}
	return out, nil
}

func parseSubPolicySpec(spec string) (EntrancePolicy, error) {
	out := EntrancePolicy{}
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		eq := strings.IndexByte(part, '=')
		if eq <= 0 {
			return EntrancePolicy{}, fmt.Errorf("expected key=value, got %q", part)
		}
		key := strings.ToLower(strings.TrimSpace(part[:eq]))
		val := strings.TrimSpace(part[eq+1:])
		switch key {
		case "uri":
			out.URI = val
		case "policy":
			if _, ok := validPolicies[val]; !ok {
				return EntrancePolicy{}, fmt.Errorf("policy=%q is not one of system|one_factor|two_factor|public", val)
			}
			out.Policy = val
		case "one_time":
			b, err := strconv.ParseBool(val)
			if err != nil {
				return EntrancePolicy{}, fmt.Errorf("one_time=%q: %w", val, err)
			}
			out.OneTime = b
		case "valid_duration":
			n, err := strconv.Atoi(val)
			if err != nil {
				return EntrancePolicy{}, fmt.Errorf("valid_duration=%q: %w", val, err)
			}
			out.ValidDuration = n
		default:
			return EntrancePolicy{}, fmt.Errorf("unknown key %q (expected uri|policy|one_time|valid_duration)", key)
		}
	}
	if out.URI == "" {
		return EntrancePolicy{}, fmt.Errorf("uri= is required")
	}
	if out.Policy == "" {
		return EntrancePolicy{}, fmt.Errorf("policy= is required")
	}
	return out, nil
}

func loadSubPoliciesFromFile(path string) ([]EntrancePolicy, error) {
	raw, err := readFileBytes(path)
	if err != nil {
		return nil, fmt.Errorf("--sub-policies-file %s: %w", path, err)
	}
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return []EntrancePolicy{}, nil
	}
	var out []EntrancePolicy
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("--sub-policies-file %s: decode JSON array: %w", path, err)
	}
	for i, e := range out {
		if e.URI == "" {
			return nil, fmt.Errorf("--sub-policies-file %s: entry[%d]: uri is required", path, i)
		}
		if _, ok := validPolicies[e.Policy]; !ok {
			return nil, fmt.Errorf("--sub-policies-file %s: entry[%d]: policy=%q is not one of system|one_factor|two_factor|public", path, i, e.Policy)
		}
	}
	return out, nil
}

func summarizeSubPolicies(p *[]EntrancePolicy) string {
	if p == nil {
		return "null"
	}
	if len(*p) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(*p))
	for _, e := range *p {
		parts = append(parts, fmt.Sprintf("%s=%s", e.URI, e.Policy))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func policyPath(app, entrance string) string {
	return "/api/applications/" + url.PathEscape(app) + "/" + url.PathEscape(entrance) + "/setup/policy"
}
