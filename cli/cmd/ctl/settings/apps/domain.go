package apps

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps domain get|set|finish <app> <entrance>`.
//
// Per-entrance custom domain editor. Mirrors the SPA's
// stores/settings/application.ts {getDomainSetup, setupDomain,
// setupCName} trio and the Settings -> App -> Entrance -> Domain
// panel (apps/.../ApplicationDomainPage.vue).
//
// Wire shape (BFL-proxied; no user-service controller):
//
//   GET  /api/applications/<app>/<entrance>/setup/domain
//        → { third_level_domain, third_party_domain,
//            cname_status, cname_target, cname_target_status,
//            cert?, key? }
//
//   POST /api/applications/<app>/<entrance>/setup/domain
//        body: { third_level_domain, third_party_domain, cert, key }
//        — replaces the current setup. Empty strings clear that
//          domain dimension. The SPA always sends ALL FOUR fields
//          (sometimes empty), so we mirror that semantic.
//
//   GET  /api/applications/<app>/<entrance>/setup/domain/finish
//        → confirms a third-party CNAME is now live (kicks the
//          server-side reconciliation that activates the cert).
//
// Set-side flag UX:
//
//   --third-level <subdomain>          set the third-level domain prefix
//   --clear-third-level                clear it
//   --third-party <fqdn>               set the third-party domain
//   --clear-third-party                clear it (incl. cert/key)
//   --cert-file <path>                 read PEM cert from file
//   --key-file <path>                  read PEM key from file
//
// We use a READ-MODIFY-WRITE pattern: every `domain set` first GETs
// the current setup so unspecified flags survive untouched. This
// matches the SPA's behavior of always preserving fields the user
// didn't change in the dialog. Without RMW, a `--third-level foo` call
// would silently zero out an existing third-party config.
//
// Role: per-app config writes; the SPA gates on isAdmin. We rely on
// server-side preflight (a normal user gets a 403 with the usual hint).

// SetupDomain mirrors the GET response payload for the
// /api/applications/<app>/<entrance>/setup/domain endpoint.
type SetupDomain struct {
	ThirdLevelDomain  string `json:"third_level_domain"`
	ThirdPartyDomain  string `json:"third_party_domain"`
	CnameStatus       string `json:"cname_status,omitempty"`
	CnameTarget       string `json:"cname_target,omitempty"`
	CnameTargetStatus string `json:"cname_target_status,omitempty"`
	Cert              string `json:"cert,omitempty"`
	Key               string `json:"key,omitempty"`
}

// setupDomainBody is the POST body shape: ALL FOUR fields, no omitempty
// — the SPA always sends them and the upstream treats absent fields as
// implicit "no change" only if cert/key are missing entirely.
type setupDomainBody struct {
	ThirdLevelDomain string `json:"third_level_domain"`
	ThirdPartyDomain string `json:"third_party_domain"`
	Cert             string `json:"cert"`
	Key              string `json:"key"`
}

// NewDomainCommand returns the `settings apps domain` parent.
func NewDomainCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "per-entrance custom domain editor (Settings -> App -> Entrance -> Domain)",
		Long: `Manage the third-level / third-party custom domain configuration for
a single app entrance. Requires both an <app> and an <entrance> name —
get the entrance names from "olares-cli settings apps entrances list".

Subcommands:
  list   <app>                                             list every entrance's domain setup
  get    <app> <entrance>                                  show current setup
  set    <app> <entrance> [--third-level X] [--third-party X.com]
                          [--cert-file PEM] [--key-file PEM]
                          [--clear-third-level] [--clear-third-party]
                                                          replace the setup (RMW)
  finish <app> <entrance>                                  confirm third-party CNAME

If you only know the app name and not its entrances yet, run
"apps domain list <app>" (or "apps entrances list <app>") first to
discover them.

Set semantics: unspecified flags survive (read-modify-write). Use
--clear-third-level / --clear-third-party to explicitly drop a
domain dimension.

Setting a third-party domain requires --cert-file AND --key-file
(unless you're using --clear-third-party). The cert/key files are
read verbatim and POSTed as multi-line strings; whatever PEM the
upstream accepts will round-trip.

After setting a third-party domain you typically need to point the
DNS CNAME at the upstream's cname_target value (visible via
"domain get") and then run "domain finish" to ask the upstream to
verify and activate the cert.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDomainListCommand(f))
	cmd.AddCommand(newDomainGetCommand(f))
	cmd.AddCommand(newDomainSetCommand(f))
	cmd.AddCommand(newDomainFinishCommand(f))
	return cmd
}

// newDomainListCommand registers `apps domain list <app>`.
//
// Convenience verb that fans out one GET /setup/domain per entrance so
// users who only know the app name don't have to run two commands
// (entrances list, then domain get per entrance). The SPA reaches the
// same effect by navigating into a per-app page that shows every
// entrance's domain inline; we mirror that flat view here. It's also
// the answer to KI-17's "apps domain get <app>" complaint — the
// previous shape required <entrance>, now the user can run "list <app>"
// to see them all at once.
func newDomainListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <app>",
		Short: "list domain setup for every entrance of an app",
		Long: `List the third-level / third-party / CNAME state for every entrance
the app exposes. Internally fans out one "domain get" per entrance and
collates the results into one table.

Pass --output json for the raw {entrance: SetupDomain} map.

Examples:
  olares-cli settings apps domain list files
  olares-cli settings apps domain list files -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDomainList(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// domainListEntry pairs an entrance name with its SetupDomain so we can
// keep the JSON output structurally explicit (callers can iterate
// entries[].entrance / entries[].setup) without us inventing an ad-hoc
// {entrance: SetupDomain} envelope.
type domainListEntry struct {
	Entrance string      `json:"entrance"`
	Setup    SetupDomain `json:"setup"`
}

func runDomainList(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
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
	entries := make([]domainListEntry, 0, len(env.Items))
	for _, e := range env.Items {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		sd, err := getDomainSetup(ctx, pc.doer, app, name)
		if err != nil {
			return fmt.Errorf("entrance %q: %w", name, err)
		}
		entries = append(entries, domainListEntry{Entrance: name, Setup: sd})
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, entries)
	}
	return renderDomainList(os.Stdout, app, entries)
}

func renderDomainList(w io.Writer, app string, entries []domainListEntry) error {
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
		if err := renderDomainSetup(w, app, e.Entrance, e.Setup); err != nil {
			return err
		}
	}
	return nil
}

func newDomainGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <app> <entrance>",
		Short: "show current per-entrance domain setup",
		Long: `Show the current third-level / third-party / CNAME state for an
entrance.

The output includes cname_status (None / Default / ThirdLevel /
ThirdParty), cname_target (the value users should CNAME their
custom domain at), and cname_target_status (whether the upstream
has detected the CNAME yet). cert / key are NEVER printed in the
default table view; pass --output json to retrieve them too.

Examples:
  olares-cli settings apps domain get files file
  olares-cli settings apps domain get files file -o json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runDomainGet(c.Context(), f, args[0], args[1], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runDomainGet(ctx context.Context, f *cmdutil.Factory, app, entrance, outputRaw string) error {
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
	current, err := getDomainSetup(ctx, pc.doer, app, entrance)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, current)
	default:
		return renderDomainSetup(os.Stdout, app, entrance, current)
	}
}

func getDomainSetup(ctx context.Context, d Doer, app, entrance string) (SetupDomain, error) {
	path := domainPath(app, entrance)
	var sd SetupDomain
	if err := doGetEnvelope(ctx, d, path, &sd); err != nil {
		return SetupDomain{}, err
	}
	return sd, nil
}

func renderDomainSetup(w io.Writer, app, entrance string, sd SetupDomain) error {
	rows := [][2]string{
		{"App", app},
		{"Entrance", entrance},
		{"Third-level domain", nonEmpty(sd.ThirdLevelDomain)},
		{"Third-party domain", nonEmpty(sd.ThirdPartyDomain)},
		{"CNAME status", nonEmpty(sd.CnameStatus)},
		{"CNAME target", nonEmpty(sd.CnameTarget)},
		{"CNAME target status", nonEmpty(sd.CnameTargetStatus)},
		{"Cert configured", boolStr(strings.TrimSpace(sd.Cert) != "")},
		{"Key configured", boolStr(strings.TrimSpace(sd.Key) != "")},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-22s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}

// newDomainSetCommand registers `apps domain set <app> <entrance>` with
// the RMW flag UX. We use explicit clear flags rather than treating an
// empty --third-level "" as "clear" because Cobra doesn't distinguish
// "flag set to empty string" from "flag absent" — both yield "" — and
// we need that distinction for the RMW merge.
func newDomainSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		thirdLevel       string
		thirdParty       string
		clearThirdLevel  bool
		clearThirdParty  bool
		certFile         string
		keyFile          string
	)
	cmd := &cobra.Command{
		Use:   "set <app> <entrance>",
		Short: "replace per-entrance domain setup (read-modify-write)",
		Long: `Replace the per-entrance custom domain configuration. Unspecified
flags survive untouched (RMW); pass --clear-third-level or
--clear-third-party to explicitly drop a domain dimension.

Setting a third-party domain requires --cert-file AND --key-file. The
files are read verbatim — pass the same PEM the upstream's UI would
accept (typically the full chain for the cert).

Examples:
  # Add a sub.example.com CNAME-style third-level domain
  olares-cli settings apps domain set files file --third-level sub

  # Switch to a fully custom third-party domain with cert/key
  olares-cli settings apps domain set files file \
      --third-party files.example.com \
      --cert-file ./fullchain.pem \
      --key-file  ./privkey.pem

  # Drop the third-party config (keeps third-level)
  olares-cli settings apps domain set files file --clear-third-party
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runDomainSet(c.Context(), f, args[0], args[1], domainSetFlags{
				thirdLevel:      thirdLevel,
				thirdLevelSet:   c.Flags().Changed("third-level"),
				thirdParty:      thirdParty,
				thirdPartySet:   c.Flags().Changed("third-party"),
				clearThirdLevel: clearThirdLevel,
				clearThirdParty: clearThirdParty,
				certFile:        certFile,
				keyFile:          keyFile,
			})
		},
	}
	cmd.Flags().StringVar(&thirdLevel, "third-level", "", "third-level subdomain prefix (e.g. \"foo\" -> foo.<userdomain>)")
	cmd.Flags().StringVar(&thirdParty, "third-party", "", "fully-qualified third-party domain (e.g. files.example.com)")
	cmd.Flags().BoolVar(&clearThirdLevel, "clear-third-level", false, "remove the third-level domain")
	cmd.Flags().BoolVar(&clearThirdParty, "clear-third-party", false, "remove the third-party domain (and its cert/key)")
	cmd.Flags().StringVar(&certFile, "cert-file", "", "path to a PEM cert file (required when setting --third-party unless --clear-third-party)")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "path to a PEM key file (required when setting --third-party unless --clear-third-party)")
	return cmd
}

type domainSetFlags struct {
	thirdLevel       string
	thirdLevelSet    bool
	thirdParty       string
	thirdPartySet    bool
	clearThirdLevel  bool
	clearThirdParty  bool
	certFile         string
	keyFile          string
}

func runDomainSet(ctx context.Context, f *cmdutil.Factory, app, entrance string, flags domainSetFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	return runDomainSetWithDoer(ctx, pc.doer, app, entrance, flags)
}

// runDomainSetWithDoer is the wire-level core of `apps domain set`. Split
// out so unit tests can drive the validation + RMW merge directly.
func runDomainSetWithDoer(ctx context.Context, d Doer, app, entrance string, flags domainSetFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app, entrance = strings.TrimSpace(app), strings.TrimSpace(entrance)
	if app == "" || entrance == "" {
		return fmt.Errorf("both <app> and <entrance> are required")
	}
	if flags.clearThirdLevel && flags.thirdLevelSet {
		return fmt.Errorf("--third-level and --clear-third-level are mutually exclusive")
	}
	if flags.clearThirdParty && flags.thirdPartySet {
		return fmt.Errorf("--third-party and --clear-third-party are mutually exclusive")
	}
	if flags.thirdPartySet && (flags.certFile == "" || flags.keyFile == "") {
		return fmt.Errorf("--third-party requires both --cert-file and --key-file")
	}
	if flags.clearThirdParty && (flags.certFile != "" || flags.keyFile != "") {
		return fmt.Errorf("--clear-third-party cannot be combined with --cert-file / --key-file")
	}
	if !flags.thirdLevelSet && !flags.clearThirdLevel &&
		!flags.thirdPartySet && !flags.clearThirdParty &&
		flags.certFile == "" && flags.keyFile == "" {
		return fmt.Errorf("nothing to do — pass at least one of --third-level / --third-party / --clear-third-level / --clear-third-party / --cert-file / --key-file")
	}
	current, err := getDomainSetup(ctx, d, app, entrance)
	if err != nil {
		return err
	}
	body := setupDomainBody{
		ThirdLevelDomain: current.ThirdLevelDomain,
		ThirdPartyDomain: current.ThirdPartyDomain,
		Cert:             current.Cert,
		Key:              current.Key,
	}
	if flags.thirdLevelSet {
		body.ThirdLevelDomain = strings.TrimSpace(flags.thirdLevel)
	}
	if flags.clearThirdLevel {
		body.ThirdLevelDomain = ""
	}
	if flags.clearThirdParty {
		body.ThirdPartyDomain = ""
		body.Cert = ""
		body.Key = ""
	}
	if flags.thirdPartySet {
		body.ThirdPartyDomain = strings.TrimSpace(flags.thirdParty)
	}
	// Cert/key updates apply alongside whatever third-party value is
	// in body now (current OR newly-set OR cleared). If --clear-third-
	// party is set we already zeroed cert/key above; the explicit-flag
	// guard rejected that combination so reaching here means we're
	// either keeping or replacing the cert/key on an active third-party.
	if flags.certFile != "" {
		certBytes, readErr := readFileBytes(flags.certFile)
		if readErr != nil {
			return fmt.Errorf("--cert-file %s: %w", flags.certFile, readErr)
		}
		body.Cert = string(certBytes)
	}
	if flags.keyFile != "" {
		keyBytes, readErr := readFileBytes(flags.keyFile)
		if readErr != nil {
			return fmt.Errorf("--key-file %s: %w", flags.keyFile, readErr)
		}
		body.Key = string(keyBytes)
	}
	if err := doMutateEnvelope(ctx, d, "POST", domainPath(app, entrance), body, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "updated domain setup for %s/%s\n", app, entrance)
	return summarizeDomainPlan(os.Stdout, body)
}

// summarizeDomainPlan renders a one-shot summary of the new setup so
// the user knows exactly what the server saw — particularly important
// because we're doing RMW merges and the user may not realize an
// existing field was preserved.
func summarizeDomainPlan(w io.Writer, body setupDomainBody) error {
	if _, err := fmt.Fprintf(w, "  third-level: %s\n", nonEmpty(body.ThirdLevelDomain)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  third-party: %s\n", nonEmpty(body.ThirdPartyDomain)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  cert:        %s\n", certKeyMark(body.Cert)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  key:         %s\n", certKeyMark(body.Key)); err != nil {
		return err
	}
	return nil
}

func certKeyMark(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(empty)"
	}
	return "(set)"
}

func newDomainFinishCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finish <app> <entrance>",
		Short: "confirm a third-party CNAME is live (Settings UI's \"Finish\" button)",
		Long: `Ask the upstream to re-check the third-party domain CNAME and finish
the domain setup. This is the same action triggered by the Settings
UI's "Finish" button on the Custom Domain dialog.

Run this AFTER you've pointed the DNS CNAME at cname_target (see
"domain get") and given DNS time to propagate. The upstream will
re-resolve the CNAME and, if it now resolves to cname_target,
activate the third-party cert and flip cname_target_status to
"completed".

Examples:
  olares-cli settings apps domain finish files file
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runDomainFinish(c.Context(), f, args[0], args[1])
		},
	}
	return cmd
}

func runDomainFinish(ctx context.Context, f *cmdutil.Factory, app, entrance string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app, entrance = strings.TrimSpace(app), strings.TrimSpace(entrance)
	if app == "" || entrance == "" {
		return fmt.Errorf("both <app> and <entrance> are required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := domainPath(app, entrance) + "/finish"
	if err := doMutateEnvelope(ctx, pc.doer, "GET", path, nil, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "submitted CNAME finalize for %s/%s\n", app, entrance)
	return nil
}

func domainPath(app, entrance string) string {
	return "/api/applications/" + url.PathEscape(app) + "/" + url.PathEscape(entrance) + "/setup/domain"
}

// readFileBytes is a thin wrapper to keep the read-pem usage testable
// (a fake doMutateEnvelope can be paired with a fake fileReader if we
// add unit tests later — for now we use the os pkg directly).
func readFileBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
