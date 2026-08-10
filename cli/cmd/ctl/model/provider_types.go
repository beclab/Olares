package model

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli model provider types [vendor|type]`
//
// GET /console/api/provider-catalog, and GET /console/api/predefined-catalog
// for --models.
//
// The catalog is what you read before `provider create`: it says which
// upstreams Router knows how to speak to, what each one calls its credential
// fields, and whether its model list is fixed or something you fill in
// yourself. It is compiled into Router — nothing here touches the network
// beyond Router itself, and nothing here reflects what is configured on this
// Olares.

type catalogI18N map[string]string

// text picks the English string, then anything, so a vendor whose YAML has no
// en_US entry still renders as itself rather than blank.
func (m catalogI18N) text() string {
	if m == nil {
		return ""
	}
	for _, k := range []string{"en_US", "en", "zh_Hans"} {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	return ""
}

type catalogFormField struct {
	Variable    string      `json:"variable"`
	Label       catalogI18N `json:"label,omitempty"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     any         `json:"default,omitempty"`
	Placeholder catalogI18N `json:"placeholder,omitempty"`
	Options     []struct {
		Value string      `json:"value"`
		Label catalogI18N `json:"label,omitempty"`
	} `json:"options,omitempty"`
	ShowOn []struct {
		Variable string `json:"variable"`
		Value    string `json:"value"`
	} `json:"show_on,omitempty"`
}

type catalogEntry struct {
	ProviderType string      `json:"provider_type"`
	Vendor       string      `json:"vendor"`
	Label        catalogI18N `json:"label"`
	Description  catalogI18N `json:"description,omitempty"`
	Help         struct {
		Title catalogI18N `json:"title,omitempty"`
		URL   catalogI18N `json:"url,omitempty"`
	} `json:"help,omitempty"`
	SupportedModelTypes []string `json:"supported_model_types"`
	ConfigurateMethods  []string `json:"configurate_methods"`
	CredentialSchema    struct {
		Fields []catalogFormField `json:"credential_form_schemas"`
	} `json:"provider_credential_schema"`
	// ModelCredentialSchema is where the vendors with no published catalog
	// declare their credential fields, because upstream Dify asks for them once
	// per model. Router stores credentials on the provider instead, so these
	// are the field names its create route wants — shown as such rather than
	// reported as "this vendor needs no credentials".
	ModelCredentialSchema *struct {
		Fields []catalogFormField `json:"credential_form_schemas"`
	} `json:"model_credential_schema,omitempty"`
	PredefinedModels []catalogModel `json:"predefined_models"`
}

// credentialFields is the field list to configure this vendor with, wherever
// the catalog happens to declare it.
func (e *catalogEntry) credentialFields() []catalogFormField {
	if len(e.CredentialSchema.Fields) > 0 {
		return e.CredentialSchema.Fields
	}
	if e.ModelCredentialSchema != nil {
		return e.ModelCredentialSchema.Fields
	}
	return nil
}

type catalogModel struct {
	Name     string          `json:"name"`
	Mode     string          `json:"mode"`
	Supports map[string]bool `json:"supports,omitempty"`
}

func newProviderTypesCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var withModels bool
	cmd := &cobra.Command{
		Use:   "types [vendor|provider-type]",
		Short: "the upstream kinds Router can speak to, and what each needs",
		Long: `List the vendors Router knows, or show one in detail.

This is Router's built-in catalog, not your configuration: it describes what
could be added, and "provider list" describes what is. Read it to learn the
provider type and credential field names that "provider create" wants.

Two words in the output decide how a provider gets its models:

  predefined-model     the vendor ships a fixed model list; you supply
                       credentials and then import the models you want
  customizable-model   you name each model yourself, because the vendor's
                       catalog is yours rather than published

A vendor can offer both. For an OpenAI-compatible endpoint that publishes its
own list — Ollama, a local model application — neither applies: use
"provider sync-models" to mirror whatever it currently serves.

With a vendor argument the credential fields are listed by their wire names,
which are exactly the keys "provider create --credential" takes. Add --models
to also list the vendor's predefined models.

Examples:
  olares-cli model provider types
  olares-cli model provider types anthropic
  olares-cli model provider types openai --models
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ref := ""
			if len(args) == 1 {
				ref = args[0]
			}
			return runProviderTypes(c.Context(), f, ref, withModels, output)
		},
	}
	cmd.Flags().BoolVar(&withModels, "models", false, "also list the vendor's predefined models (requires a vendor argument)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderTypes(ctx context.Context, f *cmdutil.Factory, ref string, withModels bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	ref = strings.TrimSpace(ref)
	if withModels && ref == "" {
		return fmt.Errorf("--models needs a vendor or provider type; run `olares-cli model provider types` to see them")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	entries, err := listProviderCatalog(ctx, pc)
	if err != nil {
		return err
	}

	if ref == "" {
		if format == FormatJSON {
			return printJSON(os.Stdout, entries)
		}
		return renderCatalogList(os.Stdout, entries)
	}

	entry, err := pickCatalogEntry(entries, ref)
	if err != nil {
		return err
	}
	var models []catalogModel
	if withModels {
		models, err = listPredefinedModels(ctx, pc, entry.ProviderType)
		if err != nil {
			return err
		}
	}
	if format == FormatJSON {
		if withModels {
			return printJSON(os.Stdout, struct {
				*catalogEntry
				PredefinedModelsResolved []catalogModel `json:"predefined_models_resolved"`
			}{entry, models})
		}
		return printJSON(os.Stdout, entry)
	}
	return renderCatalogEntry(os.Stdout, entry, withModels, models)
}

func listProviderCatalog(ctx context.Context, pc *preparedClient) ([]catalogEntry, error) {
	var env struct {
		Providers []catalogEntry `json:"providers"`
	}
	if err := pc.router.doJSON(ctx, "GET", consoleAPI+"/provider-catalog", nil, &env); err != nil {
		return nil, err
	}
	return env.Providers, nil
}

// listPredefinedModels reads the catalog by provider type. Several vendors can
// share a type — everything OpenAI-compatible that was never promoted to its
// own type lands together — so this returns the type's models rather than one
// vendor's.
func listPredefinedModels(ctx context.Context, pc *preparedClient, providerType string) ([]catalogModel, error) {
	var env struct {
		Items []catalogModel `json:"items"`
	}
	path := consoleAPI + "/predefined-catalog?provider_type=" + url.QueryEscape(providerType)
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return nil, err
	}
	return env.Items, nil
}

func pickCatalogEntry(entries []catalogEntry, ref string) (*catalogEntry, error) {
	var byType []*catalogEntry
	for i := range entries {
		if strings.EqualFold(entries[i].Vendor, ref) {
			return &entries[i], nil
		}
		if strings.EqualFold(entries[i].ProviderType, ref) {
			byType = append(byType, &entries[i])
		}
	}
	switch len(byType) {
	case 0:
		names := make([]string, 0, len(entries))
		for i := range entries {
			names = append(names, entries[i].Vendor)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("no vendor or provider type %q in Router's catalog; known vendors are %s",
			ref, strings.Join(names, ", "))
	case 1:
		return byType[0], nil
	}
	vendors := make([]string, 0, len(byType))
	for _, e := range byType {
		vendors = append(vendors, e.Vendor)
	}
	sort.Strings(vendors)
	return nil, fmt.Errorf("provider type %q covers several vendors (%s); name the one you mean",
		ref, strings.Join(vendors, ", "))
}

func renderCatalogList(w io.Writer, entries []catalogEntry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "Router reports an empty vendor catalog, which means it started without one — "+
			"check the application's logs with `olares-cli cluster logs`")
		return err
	}
	sorted := make([]catalogEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Vendor < sorted[j].Vendor })

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "VENDOR\tPROVIDER TYPE\tLABEL\tHOW MODELS ARRIVE\tSERVES\tPREDEFINED\tCREDENTIAL FIELDS"); err != nil {
		return err
	}
	for i := range sorted {
		e := &sorted[i]
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			nonEmpty(e.Vendor),
			nonEmpty(e.ProviderType),
			nonEmpty(e.Label.text()),
			nonEmpty(strings.Join(e.ConfigurateMethods, ",")),
			nonEmpty(strings.Join(e.SupportedModelTypes, ",")),
			len(e.PredefinedModels),
			nonEmpty(requiredFieldNames(e)),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\nThe credential fields column lists the required ones. "+
		"`provider types <vendor>` shows every field, including the optional ones.")
	return err
}

func requiredFieldNames(e *catalogEntry) string {
	fields := e.credentialFields()
	names := make([]string, 0, len(fields))
	for _, fld := range fields {
		if fld.Required {
			names = append(names, fld.Variable)
		}
	}
	return strings.Join(names, ",")
}

func renderCatalogEntry(w io.Writer, e *catalogEntry, withModels bool, models []catalogModel) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"VENDOR", nonEmpty(e.Vendor)},
		{"PROVIDER TYPE", nonEmpty(e.ProviderType)},
		{"LABEL", nonEmpty(e.Label.text())},
		{"HOW MODELS ARRIVE", nonEmpty(strings.Join(e.ConfigurateMethods, ", "))},
		{"SERVES", nonEmpty(strings.Join(e.SupportedModelTypes, ", "))},
	}
	if d := e.Description.text(); d != "" {
		rows = append(rows, [2]string{"DESCRIPTION", clip(d, 120)})
	}
	if u := e.Help.URL.text(); u != "" {
		rows = append(rows, [2]string{"WHERE TO GET A KEY", u})
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", r[0], r[1]); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	fields := e.credentialFields()
	if len(fields) == 0 {
		if _, err := fmt.Fprintln(w, "\nThis vendor declares no credential fields. "+
			"Send whatever the endpoint itself requires — usually api_key, or an empty object for one that needs none."); err != nil {
			return err
		}
	} else {
		if len(e.CredentialSchema.Fields) == 0 {
			if _, err := fmt.Fprintln(w, "\nThis vendor declares credentials per model rather than per provider, "+
				"and the fields below are that per-model list. Router keeps credentials on the provider and needs only "+
				"what the endpoint asks for — usually api_key; the base URL is its own flag on `provider create`."); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\nCREDENTIAL FIELDS (%d)\n", len(fields)); err != nil {
			return err
		}
		ftw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		if _, err := fmt.Fprintln(ftw, "FIELD\tKIND\tREQUIRED\tDEFAULT\tONLY WHEN\tDESCRIPTION"); err != nil {
			return err
		}
		for i := range fields {
			fld := &fields[i]
			if _, err := fmt.Fprintf(ftw, "%s\t%s\t%s\t%s\t%s\t%s\n",
				nonEmpty(fld.Variable),
				nonEmpty(fld.Type),
				boolStr(fld.Required),
				nonEmpty(formatFieldDefault(fld.Default)),
				nonEmpty(formatShowOn(fld)),
				nonEmpty(fieldDescription(fld)),
			); err != nil {
				return err
			}
		}
		if err := ftw.Flush(); err != nil {
			return err
		}
	}

	if !withModels {
		if len(e.PredefinedModels) > 0 {
			_, err := fmt.Fprintf(w, "\n%d predefined models; add --models to list them.\n", len(e.PredefinedModels))
			return err
		}
		return nil
	}
	if len(models) == 0 {
		_, err := fmt.Fprintf(w, "\nno predefined models for provider type %s — its models are yours to name, "+
			"or come from the upstream via `provider sync-models`\n", e.ProviderType)
		return err
	}
	if _, err := fmt.Fprintf(w, "\nPREDEFINED MODELS for provider type %s (%d)\n", e.ProviderType, len(models)); err != nil {
		return err
	}
	mtw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(mtw, "NAME\tMODE\tCAPABILITIES"); err != nil {
		return err
	}
	for i := range models {
		m := &models[i]
		if _, err := fmt.Fprintf(mtw, "%s\t%s\t%s\n",
			nonEmpty(m.Name), nonEmpty(m.Mode), summarizeSupports(m.Supports)); err != nil {
			return err
		}
	}
	return mtw.Flush()
}

func formatFieldDefault(v any) string {
	if v == nil {
		return ""
	}
	return clip(fmt.Sprintf("%v", v), 40)
}

// formatShowOn renders a field's visibility condition, which is the difference
// between "you must send this" and "you must send this only for one variant"
// — Azure's deployment fields and vendor-specific endpoints both work this way.
func formatShowOn(fld *catalogFormField) string {
	if len(fld.ShowOn) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fld.ShowOn))
	for _, s := range fld.ShowOn {
		parts = append(parts, s.Variable+"="+s.Value)
	}
	return strings.Join(parts, " and ")
}

func fieldDescription(fld *catalogFormField) string {
	if l := fld.Label.text(); l != "" {
		return clip(l, 48)
	}
	return clip(fld.Placeholder.text(), 48)
}
