package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router quota …` — the ceilings Router enforces.
//
// GET    /console/api/quotas
// POST   /console/api/quotas
// PUT    /console/api/quotas/:id
// DELETE /console/api/quotas/:id
//
// A quota is attached to a scope — one key, one person, one model, or one
// Olares application — and caps either spend or rate. Router evaluates them on
// the call path, so a breach is a refused request rather than a report after the
// fact.
//
// The application scope is the only control there is over an application. It
// arrives vouched for by the platform, carrying an appid rather than a
// credential Router issued, so there is nothing to revoke: a ceiling of zero is
// how one is stopped and raising it is how it starts again.
//
// The routes are per (scope, kind) rows with their own ids, which is a shape
// nobody wants to hold in their head. This tree presents them as settings on a
// thing instead: `quota set` writes the row that exists or creates the one that
// does not, and `quota clear` removes by what it caps rather than by row id.

type quotaRow struct {
	ID               int64      `json:"id"`
	ScopeType        string     `json:"scope_type"`
	ScopeID          string     `json:"scope_id"`
	QuotaType        string     `json:"quota_type"`
	LimitValue       float64    `json:"limit_value"`
	Period           string     `json:"period"`
	PeriodStart      *time.Time `json:"period_start,omitempty"`
	PeriodEnd        *time.Time `json:"period_end,omitempty"`
	SoftThresholdPct int        `json:"soft_threshold_pct"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

const (
	quotaBudget = "max_budget_usd"
	quotaRPM    = "max_rpm"
	quotaTPM    = "max_tpm"

	scopeKey   = "api_key"
	scopeUser  = "user"
	scopeModel = "provider_model"
	// scopeApp caps an Olares application by the appid its calls carry. It is
	// the only control there is over one: an application is not registered
	// with Router and has no key to revoke, so a ceiling of zero is how it is
	// stopped, and raising it is how it starts again.
	scopeApp = "app"
)

func NewQuotaCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "spend and rate ceilings on a key, a person, a model, or an application",
		Long: `Cap what a key, a person, a model or an application may consume.

Router checks these on the call path, so a breach is a refused request rather
than a line in a report. Three kinds:

  --budget  total US dollars spent, for all time
  --rpm     requests per minute
  --tpm     tokens per minute

Each is capped independently, and a request has to satisfy every quota that
covers it: a key inside its own budget is still refused when the person holding
it is over theirs.

Budget is cumulative and never resets — Router has no billing period yet. A
budget reached is a scope that stops working until the number is raised.

Subcommands:
  list           every quota, with what it applies to
  set            add or change a ceiling
  clear          remove a ceiling

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newQuotaListCommand(f))
	cmd.AddCommand(newQuotaSetCommand(f))
	cmd.AddCommand(newQuotaClearCommand(f))
	return cmd
}

// quotaTarget is the scope named on the command line, resolved to what the
// routes take.
type quotaTarget struct {
	ScopeType string
	ScopeID   string
	Label     string
}

// quotaRefs is the four ways of naming a scope, as typed. Exactly one is set.
type quotaRefs struct {
	Key   string
	User  string
	Model string
	App   string
}

func (r quotaRefs) given() int {
	n := 0
	for _, s := range []string{r.Key, r.User, r.Model, r.App} {
		if strings.TrimSpace(s) != "" {
			n++
		}
	}
	return n
}

const quotaScopeFlags = "--key, --user, --model or --caller-app"

func resolveQuotaTarget(ctx context.Context, pc *preparedClient, refs quotaRefs) (*quotaTarget, error) {
	switch refs.given() {
	case 1:
	case 0:
		return nil, fmt.Errorf("name what the quota applies to with %s", quotaScopeFlags)
	default:
		return nil, fmt.Errorf("a quota applies to one thing; pass only one of %s", quotaScopeFlags)
	}

	if s := strings.TrimSpace(refs.Key); s != "" {
		found, err := resolveKey(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		return &quotaTarget{ScopeType: scopeKey, ScopeID: found.ID,
			Label: fmt.Sprintf("key %s (%s)", found.Name, found.KeyPrefix)}, nil
	}
	if s := strings.TrimSpace(refs.User); s != "" {
		id, err := resolveUserID(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		return &quotaTarget{ScopeType: scopeUser, ScopeID: id, Label: "user " + s}, nil
	}
	if s := strings.TrimSpace(refs.App); s != "" {
		id, err := resolveCallerAppID(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		return &quotaTarget{ScopeType: scopeApp, ScopeID: id, Label: "application " + s}, nil
	}
	row, err := resolveModel(ctx, pc, strings.TrimSpace(refs.Model))
	if err != nil {
		return nil, err
	}
	return &quotaTarget{ScopeType: scopeModel, ScopeID: row.ProviderModelID,
		Label: "model " + row.label()}, nil
}

func newQuotaListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		refs   quotaRefs
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "every quota, with what it applies to",
		Long: `List the quotas in force.

APPLIES TO names the key, person, model or application each one covers, rather
than the id Router stores, because an id says nothing about who is about to be
refused.

Narrow to one scope with --key, --user, --model or --caller-app.

Examples:
  olares-cli router quota list
  olares-cli router quota list --user pptest01
  olares-cli router quota list -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runQuotaList(c.Context(), f, refs, output)
		},
	}
	addQuotaScopeFlags(cmd, &refs, "only quotas on")
	addOutputFlag(cmd, &output)
	return cmd
}

// addQuotaScopeFlags names the four scopes the same way in all three verbs.
// lead is the clause each usage line starts with: "the key to cap", "only
// quotas on this key".
func addQuotaScopeFlags(cmd *cobra.Command, refs *quotaRefs, lead string) {
	cmd.Flags().StringVar(&refs.Key, "key", "", lead+" this key, by name, prefix or id")
	cmd.Flags().StringVar(&refs.User, "user", "", lead+" this user, by name or id")
	cmd.Flags().StringVar(&refs.Model, "model", "", lead+" this model, as <provider>/<model>")
	cmd.Flags().StringVar(&refs.App, "caller-app", "", lead+" this application, by title, app name or appid")
}

func runQuotaList(ctx context.Context, f *cmdutil.Factory, refs quotaRefs, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var target *quotaTarget
	if refs.given() > 0 {
		target, err = resolveQuotaTarget(ctx, pc, refs)
		if err != nil {
			return err
		}
	}
	rows, err := listQuotas(ctx, pc, target)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": rows})
	}
	return renderQuotaList(ctx, pc, os.Stdout, rows, target)
}

func listQuotas(ctx context.Context, pc *preparedClient, target *quotaTarget) ([]quotaRow, error) {
	q := url.Values{}
	if target != nil {
		q.Set("scope_type", target.ScopeType)
		q.Set("scope_id", target.ScopeID)
	}
	return collection[quotaRow](ctx, pc, withQuery(epQuotas, q))
}

func renderQuotaList(ctx context.Context, pc *preparedClient, w io.Writer, rows []quotaRow, target *quotaTarget) error {
	if len(rows) == 0 {
		what := "no quotas are set, so nothing is capped"
		if target != nil {
			what = fmt.Sprintf("no quota applies to %s", target.Label)
		}
		_, err := fmt.Fprintf(w, "%s. `olares-cli router quota set --user <name> --budget 25` is the shape of one.\n", what)
		return err
	}
	labels := scopeLabels(ctx, pc, rows)
	t := newTable(w, "APPLIES TO", "CAPS", "LIMIT", "WARN AT", "ID")
	for i := range rows {
		r := &rows[i]
		label := labels[r.ScopeType+":"+r.ScopeID]
		if label == "" {
			label = r.ScopeType + " " + r.ScopeID
		}
		t.row(label, quotaKindLabel(r.QuotaType), quotaLimitLabel(r),
			fmt.Sprintf("%d%%", r.SoftThresholdPct), strconv.FormatInt(r.ID, 10))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\nWARN AT is where Router starts warning; the request is only refused at the limit. "+
		"A budget is cumulative and does not reset.")
	return err
}

func quotaKindLabel(t string) string {
	switch t {
	case quotaBudget:
		return "spend"
	case quotaRPM:
		return "requests/min"
	case quotaTPM:
		return "tokens/min"
	default:
		return t
	}
}

func quotaLimitLabel(r *quotaRow) string {
	if r.QuotaType == quotaBudget {
		return "$" + strconv.FormatFloat(r.LimitValue, 'f', -1, 64)
	}
	return strconv.FormatFloat(r.LimitValue, 'f', -1, 64)
}

// scopeLabels names the things quotas point at. Each list is read at most once,
// and only for the scope kinds actually present.
func scopeLabels(ctx context.Context, pc *preparedClient, rows []quotaRow) map[string]string {
	out := map[string]string{}
	need := map[string]bool{}
	for i := range rows {
		need[rows[i].ScopeType] = true
	}
	if need[scopeKey] {
		if keys, err := listKeys(ctx, pc); err == nil {
			for i := range keys {
				out[scopeKey+":"+keys[i].ID] = fmt.Sprintf("key %s (%s)", keys[i].Name, keys[i].KeyPrefix)
			}
		}
	}
	if need[scopeUser] {
		for id, name := range userLabels(ctx, pc) {
			out[scopeUser+":"+id] = "user " + name
		}
	}
	if need[scopeModel] {
		if models, err := listAllModels(ctx, pc); err == nil {
			for i := range models {
				m := &models[i]
				out[scopeModel+":"+m.Model.ID] = fmt.Sprintf("model %s on %s", m.Model.Name, m.ProviderName)
			}
		}
	}
	if need[scopeApp] {
		if rows, err := callerAppBuckets(ctx, pc); err == nil {
			for i := range rows {
				out[scopeApp+":"+rows[i].Key] = "application " + nonEmpty(rows[i].Label)
			}
		}
	}
	return out
}

func newQuotaSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		refs   quotaRefs
		budget float64
		rpm    float64
		tpm    float64
		warnAt int
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "add or change a ceiling",
		Long: `Set a quota on a key, a person, a model, or an application.

Name the scope with exactly one of --key, --user, --model or --caller-app, and
what to cap with one or more of --budget, --rpm and --tpm. Each kind is a
separate ceiling, so setting two in one command writes two of them.

An existing ceiling of the same kind is changed rather than duplicated, which is
what makes this safe to run twice.

--warn-at moves the point where Router starts warning, as a percentage of the
limit. It defaults to 80 on a new quota and is left alone on an existing one
unless given.

Raising a budget is how a scope that has stopped working starts again: spend is
cumulative and nothing resets it.

--caller-app caps an Olares application. It is the only control over one: an
application is vouched for by the platform rather than registered here, so there
is no key of its own to revoke, and a ceiling is what stops it.

Examples:
  olares-cli router quota set --user pptest01 --budget 25
  olares-cli router quota set --key ci --rpm 60 --tpm 120000
  olares-cli router quota set --model "openai/gpt-4o" --rpm 10 --warn-at 90
  olares-cli router quota set --caller-app wise --budget 5
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			want := map[string]float64{}
			if c.Flags().Changed("budget") {
				want[quotaBudget] = budget
			}
			if c.Flags().Changed("rpm") {
				want[quotaRPM] = rpm
			}
			if c.Flags().Changed("tpm") {
				want[quotaTPM] = tpm
			}
			if len(want) == 0 {
				return fmt.Errorf("say what to cap with --budget, --rpm or --tpm")
			}
			for kind, v := range want {
				if v <= 0 {
					return fmt.Errorf("--%s must be greater than zero; "+
						"`olares-cli router quota clear` removes a ceiling instead", flagForQuotaKind(kind))
				}
			}
			var pct *int
			if c.Flags().Changed("warn-at") {
				if warnAt < 0 || warnAt > 100 {
					return fmt.Errorf("--warn-at is a percentage between 0 and 100, not %d", warnAt)
				}
				pct = &warnAt
			}
			return runQuotaSet(c.Context(), f, refs, want, pct, output)
		},
	}
	addQuotaScopeFlags(cmd, &refs, "cap")
	cmd.Flags().Float64Var(&budget, "budget", 0, "cap total spend, in US dollars, for all time")
	cmd.Flags().Float64Var(&rpm, "rpm", 0, "cap requests per minute")
	cmd.Flags().Float64Var(&tpm, "tpm", 0, "cap tokens per minute")
	cmd.Flags().IntVar(&warnAt, "warn-at", 80, "warn at this percentage of the limit")
	addOutputFlag(cmd, &output)
	return cmd
}

func flagForQuotaKind(kind string) string {
	switch kind {
	case quotaBudget:
		return "budget"
	case quotaRPM:
		return "rpm"
	default:
		return "tpm"
	}
}

func runQuotaSet(ctx context.Context, f *cmdutil.Factory, refs quotaRefs, want map[string]float64, warnAt *int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	target, err := resolveQuotaTarget(ctx, pc, refs)
	if err != nil {
		return err
	}
	existing, err := listQuotas(ctx, pc, target)
	if err != nil {
		return err
	}
	byKind := map[string]*quotaRow{}
	for i := range existing {
		byKind[existing[i].QuotaType] = &existing[i]
	}

	kinds := make([]string, 0, len(want))
	for k := range want {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)

	written := make([]quotaRow, 0, len(kinds))
	for _, kind := range kinds {
		limit := want[kind]
		var row quotaRow
		if cur, ok := byKind[kind]; ok {
			body := map[string]any{"limit_value": limit}
			if warnAt != nil {
				body["soft_threshold_pct"] = *warnAt
			}
			path := epQuota(cur.ID)
			if err := pc.router.doJSON(ctx, "PUT", path, body, &row); err != nil {
				return err
			}
		} else {
			body := map[string]any{
				"scope_type":  target.ScopeType,
				"scope_id":    target.ScopeID,
				"quota_type":  kind,
				"limit_value": limit,
			}
			if warnAt != nil {
				body["soft_threshold_pct"] = *warnAt
			}
			if err := pc.router.doJSON(ctx, "POST", epQuotas, body, &row); err != nil {
				return err
			}
		}
		written = append(written, row)
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": written})
	}
	for i := range written {
		r := &written[i]
		if _, err := fmt.Fprintf(os.Stdout, "%s is capped at %s %s, warning from %d%%\n",
			target.Label, quotaLimitLabel(r), quotaKindLabel(r.QuotaType), r.SoftThresholdPct); err != nil {
			return err
		}
	}
	return nil
}

func newQuotaClearCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		refs      quotaRefs
		budget    bool
		rpm       bool
		tpm       bool
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "remove a ceiling",
		Long: `Remove quotas from a key, a person, a model, or an application.

Without --budget, --rpm or --tpm every ceiling on the scope goes, which is the
usual intent and the reason for the prompt. Naming kinds removes only those.

Removing a ceiling is not the same as raising it: nothing caps the scope
afterwards.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Examples:
  olares-cli router quota clear --key ci --yes
  olares-cli router quota clear --user pptest01 --budget
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			kinds := map[string]bool{}
			if budget {
				kinds[quotaBudget] = true
			}
			if rpm {
				kinds[quotaRPM] = true
			}
			if tpm {
				kinds[quotaTPM] = true
			}
			return runQuotaClear(c.Context(), f, refs, kinds, assumeYes)
		},
	}
	addQuotaScopeFlags(cmd, &refs, "uncap")
	cmd.Flags().BoolVar(&budget, "budget", false, "remove only the spend ceiling")
	cmd.Flags().BoolVar(&rpm, "rpm", false, "remove only the requests-per-minute ceiling")
	cmd.Flags().BoolVar(&tpm, "tpm", false, "remove only the tokens-per-minute ceiling")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	return cmd
}

func runQuotaClear(ctx context.Context, f *cmdutil.Factory, refs quotaRefs, kinds map[string]bool, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	target, err := resolveQuotaTarget(ctx, pc, refs)
	if err != nil {
		return err
	}
	existing, err := listQuotas(ctx, pc, target)
	if err != nil {
		return err
	}
	doomed := make([]quotaRow, 0, len(existing))
	for i := range existing {
		if len(kinds) == 0 || kinds[existing[i].QuotaType] {
			doomed = append(doomed, existing[i])
		}
	}
	if len(doomed) == 0 {
		_, werr := fmt.Fprintf(os.Stdout, "nothing to remove: no such quota applies to %s\n", target.Label)
		return werr
	}

	if !assumeYes {
		what := make([]string, 0, len(doomed))
		for i := range doomed {
			what = append(what, fmt.Sprintf("%s %s", quotaLimitLabel(&doomed[i]), quotaKindLabel(doomed[i].QuotaType)))
		}
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Remove the %s ceiling on %s? Nothing will cap it afterwards.",
				strings.Join(what, " and "), target.Label),
			false); err != nil {
			return err
		}
	}
	for i := range doomed {
		path := epQuota(doomed[i].ID)
		if err := pc.router.doJSON(ctx, "DELETE", path, nil, nil); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "removed the %s ceiling on %s\n",
			quotaKindLabel(doomed[i].QuotaType), target.Label); err != nil {
			return err
		}
	}
	return nil
}
