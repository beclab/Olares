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
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router list` — GET /console/api/provider-models
// `olares-cli router capabilities` — GET /console/api/capabilities/supports
//
// One request answers "what can this Olares call", across every provider. It is
// the only read in this tree open to a non-admin user, which is deliberate: the
// list carries model and provider metadata and never a credential, and choosing
// which models a key may reach is something a user does for themselves.
//
// A model appearing here is not the same as a model being callable. The row says
// what is configured; whether the upstream answers is what `provider validate`
// and `router call` find out.

// adminModelRow is one model with its provider lifted alongside it, which is
// what makes a flat list readable: the same model name can exist on two
// providers, and the provider is what tells them apart.
type adminModelRow struct {
	ProviderModelID string           `json:"provider_model_id"`
	ProviderID      string           `json:"provider_id"`
	ProviderName    string           `json:"provider_name"`
	ProviderType    string           `json:"provider_type"`
	ProviderSource  string           `json:"provider_source"`
	ProviderStatus  string           `json:"provider_status"`
	Model           providerModelRow `json:"model"`
}

func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		providerRef  string
		mode         string
		source       string
		status       string
		enabledOnly  bool
		disabledOnly bool
		search       string
		limit        int
		offset       int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "every model configured across every provider",
		Long: `List the models Router has configured, across all providers.

The same model name can exist on more than one provider, so the provider column
is part of the identity rather than decoration.

Two columns are easy to conflate. "OFFERED" is whether the model is handed to
callers at all; "STATUS" is the row's own state. A model can be present,
active, and still not offered, which is what "models update --disable" leaves
behind.

What a model claims to support, its context window and its prices are not in
this answer — Router keeps them out of the aggregate list. "router provider get
<provider>" carries them.

Being listed does not mean being reachable. This is the configuration; whether
the upstream answers right now is what "provider validate" reports.

Results are capped, newest first. Narrow with --provider, --mode, --search or
--enabled rather than raising --limit when you are looking for one model.

This is the one read here that does not require an admin session: no credential
appears in it, and a user needs it to choose which models their own key may
reach.

Examples:
  olares-cli router list
  olares-cli router list --mode embedding
  olares-cli router list --provider claude --enabled
  olares-cli router list --search qwen -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if enabledOnly && disabledOnly {
				return fmt.Errorf("pass either --enabled or --disabled, not both")
			}
			filter := modelListFilter{
				ProviderRef: strings.TrimSpace(providerRef),
				Mode:        strings.ToLower(strings.TrimSpace(mode)),
				Source:      strings.ToLower(strings.TrimSpace(source)),
				Status:      strings.ToLower(strings.TrimSpace(status)),
				Search:      strings.TrimSpace(search),
				Limit:       limit,
				Offset:      offset,
			}
			if enabledOnly {
				v := true
				filter.Enabled = &v
			}
			if disabledOnly {
				v := false
				filter.Enabled = &v
			}
			return runModelList(c.Context(), f, filter, output)
		},
	}
	cmd.Flags().StringVar(&providerRef, "provider", "", "only models on this provider, by name or id")
	cmd.Flags().StringVar(&mode, "mode", "", "only this endpoint family: "+strings.Join(providerModelModes, ", "))
	cmd.Flags().StringVar(&source, "source", "", "manual or olares")
	cmd.Flags().StringVar(&status, "status", "", "active or disabled")
	cmd.Flags().BoolVar(&enabledOnly, "enabled", false, "only models offered to callers")
	cmd.Flags().BoolVar(&disabledOnly, "disabled", false, "only models not offered to callers")
	cmd.Flags().StringVar(&search, "search", "", "match part of a model or provider name")
	cmd.Flags().IntVar(&limit, "limit", 100, "how many rows to return (1-1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "how many rows to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

type modelListFilter struct {
	ProviderRef string
	Mode        string
	Source      string
	Status      string
	Enabled     *bool
	Search      string
	Limit       int
	Offset      int
}

func runModelList(ctx context.Context, f *cmdutil.Factory, filter modelListFilter, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if filter.Mode != "" && !containsString(providerModelModes, filter.Mode) {
		return fmt.Errorf("--mode must be one of %s, not %q", strings.Join(providerModelModes, ", "), filter.Mode)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	q := url.Values{}
	if filter.ProviderRef != "" {
		// The route filters by provider id, so a name has to be resolved first.
		found, rerr := resolveProvider(ctx, pc, filter.ProviderRef)
		if rerr != nil {
			return rerr
		}
		q.Set("provider_id", found.ID)
	}
	if filter.Mode != "" {
		q.Set("mode", filter.Mode)
	}
	if filter.Source != "" {
		q.Set("source", filter.Source)
	}
	if filter.Status != "" {
		q.Set("status", filter.Status)
	}
	if filter.Enabled != nil {
		q.Set("enabled", strconv.FormatBool(*filter.Enabled))
	}
	if filter.Search != "" {
		q.Set("search", filter.Search)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}
	if filter.Offset > 0 {
		q.Set("offset", strconv.Itoa(filter.Offset))
	}

	var env struct {
		Items  []adminModelRow `json:"items"`
		Total  int             `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}
	path := consoleAPI + "/provider-models"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderModelList(os.Stdout, env.Items, env.Total, env.Limit, env.Offset)
}

func renderModelList(w io.Writer, items []adminModelRow, total, limit, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no models match. If nothing at all is configured, start with "+
			"`olares-cli router provider types` to see what can be added.")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "MODEL\tPROVIDER\tTYPE\tSOURCE\tMODE\tOFFERED\tSTATUS\tPROVIDER STATUS"); err != nil {
		return err
	}
	for i := range items {
		it := &items[i]
		name := it.Model.Name
		if it.Model.Alias != nil && *it.Model.Alias != "" {
			name = fmt.Sprintf("%s (as %s)", it.Model.Name, *it.Model.Alias)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(name),
			nonEmpty(it.ProviderName),
			nonEmpty(it.ProviderType),
			nonEmpty(it.ProviderSource),
			nonEmpty(it.Model.Mode),
			boolStr(it.Model.Enabled),
			nonEmpty(it.Model.Status),
			nonEmpty(it.ProviderStatus),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	shown := offset + len(items)
	if total > shown {
		_, err := fmt.Fprintf(w, "\nshowing %d-%d of %d; --offset %d for the next page\n",
			offset+1, shown, total, shown)
		return err
	}
	return nil
}

// `router capabilities` is the vocabulary for --supports. It is a fixed list
// compiled into Router, so it answers "what may I claim" rather than "what does
// anything support".
func NewCapabilitiesCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "the capability flags a model row can declare",
		Long: `List every capability flag Router understands.

These are the keys "provider models add --supports" and "provider models update
--supports" accept. The list is fixed in the Router build you are talking to, so
a flag missing here is not one this Router will honour.

Router checks these flags before dispatching a request, so they are a promise
rather than a description: a model whose vision flag is unset will have image
requests refused even if the upstream would have accepted them, and one whose
flag is set wrongly will forward requests the upstream then rejects.

Example:
  olares-cli router capabilities
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCapabilities(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runCapabilities(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var env struct {
		Supports []string `json:"supports"`
	}
	if err := pc.router.doJSON(ctx, "GET", consoleAPI+"/capabilities/supports", nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	sorted := make([]string, len(env.Supports))
	copy(sorted, env.Supports)
	sort.Strings(sorted)
	for _, s := range sorted {
		if _, err := fmt.Fprintln(os.Stdout, s); err != nil {
			return err
		}
	}
	return nil
}
