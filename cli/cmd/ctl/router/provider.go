package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider` — the connections Router routes through.
//
// A provider is one upstream that serves models: a cloud account (OpenAI,
// Anthropic, Gemini, Azure, or anything OpenAI-compatible) or a model
// application installed on this Olares. Both kinds are rows in the same
// table, and `source` is what separates them: `manual` for one an admin
// entered, `olares` for one Router registered itself when a model app
// appeared. That distinction decides what may be edited — a Market-sourced
// provider is owned by its app, so its credentials and lifecycle belong to
// the Market verbs rather than to `provider update`.
//
// Every route under this parent is admin-only.

// providerRow is Router's safe provider projection. Credentials never appear
// on the wire, by three separate guards in Router; there is nothing to
// redact here.
type providerRow struct {
	ID string `json:"id"`
	// Name is what a caller writes in front of a model name, and it is not
	// unique: every locally installed model application carries the one name
	// `Olares`, so that a caller writes `Olares/<model>` without having to
	// know which application serves it. Which application a row is stays in
	// OlaresAppName, and what an operator called it in ProviderDisplayTitle.
	Name                 string  `json:"name"`
	ProviderDisplayTitle *string `json:"provider_display_title,omitempty"`
	ProviderType         string  `json:"provider_type"`
	BaseURL              string  `json:"base_url"`
	Status               string  `json:"status"`
	Source               string  `json:"source"`
	OlaresAppName        *string `json:"olares_app_name,omitempty"`
	OlaresVersionName    *string `json:"olares_version_name,omitempty"`
	OlaresStatus         *string `json:"olares_status,omitempty"`
	// OlaresMarketSource is the Market source the installed copy came from.
	// Absent means the copy predates Router recording it, which reads as
	// "cannot tell" rather than as a mismatch.
	OlaresMarketSource *string `json:"olares_market_source,omitempty"`
	// ModelConsoleStatus is the phase the application's own control plane
	// reports for the model it serves, alongside olares_status: that is the
	// container's lifecycle, this is the model's. Present on every locally
	// installed application, absent on one an admin entered.
	ModelConsoleStatus *string   `json:"model_console_status,omitempty"`
	OlaresLatestTaskID *string   `json:"olares_latest_task_id,omitempty"`
	CredentialsVersion int       `json:"credentials_version"`
	IconURL            *string   `json:"icon_url,omitempty"`
	EntranceURL        *string   `json:"entrance_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// providerModelRow is one model a provider serves. The capability fields are
// flat siblings in Router's v2 model schema; `supports` is a wide map of
// booleans and `pricing` a wide map of decimal strings, so both are rendered
// selectively rather than dumped in table form.
//
// ParameterRules is kept as raw JSON. Nothing here reads inside it, and it has
// to survive an edit to a neighbouring field intact — see the read-modify-write
// in `models update`.
type providerModelRow struct {
	ID              string            `json:"id"`
	ProviderID      string            `json:"provider_id"`
	Name            string            `json:"name"`
	Mode            string            `json:"mode"`
	Enabled         bool              `json:"enabled"`
	Status          string            `json:"status"`
	Supports        map[string]bool   `json:"supports,omitempty"`
	Pricing         map[string]string `json:"pricing,omitempty"`
	ParameterRules  json.RawMessage   `json:"parameter_rules,omitempty"`
	ContextSize     int               `json:"context_size,omitempty"`
	MaxOutputTokens int               `json:"max_output_tokens,omitempty"`
	EngineArgs      string            `json:"engine_args,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type providerDetail struct {
	providerRow
	Models []providerModelRow `json:"models"`
}

func (p *providerRow) title() string {
	if p.ProviderDisplayTitle != nil && *p.ProviderDisplayTitle != "" {
		return *p.ProviderDisplayTitle
	}
	if p.OlaresAppName != nil && *p.OlaresAppName != "" {
		return *p.OlaresAppName
	}
	return p.Name
}

// handle is what to print when a message asks someone to name this provider
// again. For a hand-entered one that is its name; for a model application it is
// the application's, because `Olares` names all of them at once and would send
// the reader back with an ambiguous argument.
func (p *providerRow) handle() string {
	if p.OlaresAppName != nil && *p.OlaresAppName != "" {
		return *p.OlaresAppName
	}
	return p.Name
}

// isMarketSourced reports whether the provider belongs to an installed model
// application. Router refuses edits and deletes on these, pointing at the
// Market flow instead, so verbs check it before sending a doomed request.
func (p *providerRow) isMarketSourced() bool {
	return strings.EqualFold(p.Source, "olares")
}

func NewProviderCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "the upstreams Router routes to (cloud accounts and local model apps)",
		Long: `Inspect the providers Router can route to.

A provider is one upstream serving models. Two kinds share the table, told
apart by "source":

  manual   an admin entered it: base URL plus credentials for OpenAI,
           Anthropic, Gemini, Azure, or any OpenAI-compatible endpoint
  olares   Router registered it itself when a model application was
           installed on this Olares; the app owns its lifecycle, so
           credentials and removal go through the Market verbs

Subcommands:
  list             every provider, newest first
  get <name>       one provider with the models it serves
  types            the vendors you can add, and the credentials each needs
  create           add a provider from a base URL and credentials
  update <name>    rename, re-point, disable, or rotate credentials
  delete <name>    remove a provider and everything hanging off it
  validate <name>  ask the upstream whether the stored credentials work
  sync-models      mirror an upstream's live model list into Router
  credentials      the stored credential fields, with secrets masked
  history <name>   credential versions, newest first
  rollback         restore an earlier credential version
  register <app>   the provider Router registered for a model application
  models           which models a provider offers

Market-sourced providers only appear in the list once their application is
actually running, so a model app that is still installing is absent here
rather than shown broken.

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newProviderListCommand(f))
	cmd.AddCommand(newProviderGetCommand(f))
	cmd.AddCommand(newProviderTypesCommand(f))
	cmd.AddCommand(newProviderCreateCommand(f))
	cmd.AddCommand(newProviderUpdateCommand(f))
	cmd.AddCommand(newProviderDeleteCommand(f))
	cmd.AddCommand(newProviderValidateCommand(f))
	cmd.AddCommand(newProviderSyncModelsCommand(f))
	cmd.AddCommand(newProviderCredentialsCommand(f))
	cmd.AddCommand(newProviderHistoryCommand(f))
	cmd.AddCommand(newProviderRollbackCommand(f))
	cmd.AddCommand(newProviderRegisterCommand(f))
	cmd.AddCommand(newProviderModelsCommand(f))
	return cmd
}

// renderProviderRow is the key-value block every verb that returns a provider
// prints. Router's response to a write is the same projection the detail route
// serves, so the reader sees the row as it now stands rather than an echo of
// what was sent.
func renderProviderRow(w io.Writer, p *providerRow) error {
	t := newTable(w)
	t.row("NAME", nonEmpty(p.Name))
	t.row("TITLE", nonEmpty(p.title()))
	t.row("TYPE", nonEmpty(p.ProviderType))
	t.row("STATUS", nonEmpty(p.Status))
	t.row("SOURCE", nonEmpty(p.Source))
	t.row("BASE URL", nonEmpty(p.BaseURL))
	t.row("CREDENTIALS VERSION", strconv.Itoa(p.CredentialsVersion))
	t.row("ID", nonEmpty(p.ID))
	if p.OlaresAppName != nil {
		t.row("OLARES APP", nonEmpty(*p.OlaresAppName))
	}
	if p.OlaresVersionName != nil {
		t.row("OLARES VERSION", nonEmpty(*p.OlaresVersionName))
	}
	if p.OlaresStatus != nil {
		t.row("OLARES STATUS", nonEmpty(*p.OlaresStatus))
	}
	if p.EntranceURL != nil {
		t.row("ENTRANCE URL", nonEmpty(*p.EntranceURL))
	}
	return t.flush()
}

// marketOwnedErr is the refusal every write verb owes a Market-sourced
// provider. Router would answer 409 anyway, but the app's own name is what
// makes the message actionable and the CLI has it in hand.
func marketOwnedErr(p *providerRow, action string) error {
	app := "its model application"
	if p.OlaresAppName != nil && *p.OlaresAppName != "" {
		app = "the " + *p.OlaresAppName + " application"
	}
	return fmt.Errorf("%q was registered by %s, so Router owns its configuration; "+
		"cannot %s it here — manage the application instead with `olares-cli market`",
		p.Name, app, action)
}

// resolveProvider turns what a person typed into a provider row.
//
// Three handles reach a row, because Router's own name for one is not always
// something a person can name it by: the routing name, the application name for
// a locally installed model application, and the display title an operator gave
// it. The routing name is not unique — every model application carries `Olares`
// — so it is matched last and an ambiguous match is refused with the names that
// would separate the rows rather than resolved to whichever came back first.
//
// An id is fetched directly rather than looked up in the list, because the
// list omits a Market-sourced provider whose application is not running while
// the detail route still serves it. That is the one way to inspect a provider
// belonging to a stopped model app.
func resolveProvider(ctx context.Context, pc *preparedClient, ref string) (*providerRow, error) {
	ref, err := requireRef(ref, "a provider name, application name or id")
	if err != nil {
		return nil, err
	}
	if entityID.MatchString(ref) {
		detail, err := getProvider(ctx, pc, ref)
		if err != nil {
			return nil, err
		}
		row := detail.providerRow
		return &row, nil
	}
	rows, err := listProviders(ctx, pc)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if strDeref(rows[i].OlaresAppName) != "" && strings.EqualFold(strDeref(rows[i].OlaresAppName), ref) {
			return &rows[i], nil
		}
	}
	for i := range rows {
		if title := strDeref(rows[i].ProviderDisplayTitle); title != "" && strings.EqualFold(title, ref) {
			return &rows[i], nil
		}
	}
	byName := make([]*providerRow, 0, 2)
	for i := range rows {
		if strings.EqualFold(rows[i].Name, ref) {
			byName = append(byName, &rows[i])
		}
	}
	if len(byName) == 1 {
		return byName[0], nil
	}
	if len(byName) > 1 {
		handles := make([]string, 0, len(byName))
		for _, row := range byName {
			handles = append(handles, row.handle())
		}
		sort.Strings(handles)
		return nil, fmt.Errorf("%q names %d providers, because every locally installed model application "+
			"answers to it; name one of them instead: %s",
			ref, len(byName), strings.Join(handles, ", "))
	}
	// The provider list drops a Market-sourced provider whose application is
	// not running, while the aggregate model list keeps it. Asking there
	// second is what lets a name — the only handle anyone has — reach a
	// provider that exists but is hidden.
	if id := providerIDFromModels(ctx, pc, ref); id != "" {
		detail, derr := getProvider(ctx, pc, id)
		if derr == nil {
			row := detail.providerRow
			return &row, nil
		}
	}
	// An application still installing is hidden from both of those: it is not
	// running, and it has never been asked for its models. The model-app list
	// is the one place that names the row anyway, which matters because the
	// minutes right after an install are when someone is most likely to name it.
	if id := providerIDFromModelApps(ctx, pc, ref); id != "" {
		detail, derr := getProvider(ctx, pc, id)
		if derr == nil {
			row := detail.providerRow
			return &row, nil
		}
	}
	known := make([]string, 0, len(rows))
	for i := range rows {
		known = append(known, rows[i].handle())
	}
	sort.Strings(known)
	return nil, missing{
		noun:  "provider",
		ref:   ref,
		known: known,
		have:  "the listed ones are",
		none:  "no provider is currently listed",
		note:  hiddenProviderNote,
	}.err()
}

// providerIDFromModels finds a provider id by name in the aggregate model list.
// Empty when the name is unknown there too, or when the list cannot be read —
// this is a fallback, and its failure should read as "not found" rather than
// replace the caller's error with one about a different route.
func providerIDFromModels(ctx context.Context, pc *preparedClient, name string) string {
	rows, err := listAllModels(ctx, pc)
	if err != nil {
		return ""
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ProviderName, name) {
			return rows[i].ProviderID
		}
	}
	return ""
}

// hiddenProviderNote explains the one case where a provider exists but is not
// listed, which is otherwise indistinguishable from having none.
const hiddenProviderNote = "A provider belonging to a model application is listed only while that application " +
	"runs or is being installed, so check `olares-cli market status <app>` for one that is stopped; " +
	"either way it can still be named by its application name or inspected by id."

func listProviders(ctx context.Context, pc *preparedClient) ([]providerRow, error) {
	return collection[providerRow](ctx, pc, epProviders)
}

func getProvider(ctx context.Context, pc *preparedClient, id string) (*providerDetail, error) {
	var detail providerDetail
	path := epProvider(id)
	if err := pc.router.doJSON(ctx, "GET", path, nil, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}
