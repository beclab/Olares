package middleware

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewListCommand: `olares-cli cluster middleware list [-t TYPE]
// [--show-passwords] [-o table|json]`.
//
// Calls SPA's getMiddlewareListAll
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:624):
// `/middleware/v1/list`. Returns `{code, data: [MiddlewareItem]}`
// — a non-K8s envelope. We unwrap and decode straight into our
// typed slice.
//
// --type filters by middleware type (mongodb / postgres / redis /
// rabbitmq / minio / nats / mysql / mariadb / elasticsearch).
// Filtering happens client-side because the per-type endpoint
// (/middleware/v1/<type>/list) returns the same shape and supports
// a different scope contract — keeping the filter local makes the
// flag a pure projection over the all-types result.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		typeFilter   string
		showPassword bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Olares middleware instances visible to the active profile",
		Long: `List Olares middleware instances visible to the active profile.

Output (table mode): TYPE, NAME, NAMESPACE, NODES, ADMIN-USER.
Admin password is never printed in table mode. In ` + "`-o json`" + ` mode,
the password is redacted as "<redacted>" unless --show-passwords is
explicitly set.

--type filters the output to one middleware type (case-insensitive).
Empty (default) means all types.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), o, typeFilter, showPassword)
		},
	}
	cmd.Flags().StringVarP(&typeFilter, "type", "t", "", "filter by middleware type (mongodb / postgres / redis / rabbitmq / minio / nats / mysql / mariadb / elasticsearch)")
	cmd.Flags().BoolVar(&showPassword, "show-passwords", false, "include the admin password in -o json output (table mode always redacts)")
	o.AddOutputFlags(cmd)
	return cmd
}

// envelope is the /middleware/v1/list response wrapper.
//
// We model only Code + Data here; the SPA also reads `data` as
// either an empty array or a real slice depending on whether any
// instances exist (see MiddlewareListResponse in
// apps/packages/app/src/apps/controlPanelCommon/network/middleware.ts).
type envelope struct {
	Code int    `json:"code"`
	Data []Item `json:"data"`
	// Some upstream paths surface a message field on error; we read
	// it best-effort to enrich the wrapped error.
	Message string `json:"message,omitempty"`
}

// Item is the per-instance projection. Mirrors MiddlewareItem in
// apps/packages/app/src/apps/controlPanelCommon/network/middleware.ts
// minus the SPA-only chip data (mongos / redisProxy / proxy
// endpoints are kept verbatim for `-o json` consumers but not
// rendered in the table).
type Item struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Nodes     int    `json:"nodes"`
	AdminUser string `json:"adminUser"`
	Password  string `json:"password,omitempty"`
	Mongos    struct {
		Endpoint string `json:"endpoint,omitempty"`
		Size     int    `json:"size,omitempty"`
	} `json:"mongos,omitempty"`
	RedisProxy struct {
		Endpoint string `json:"endpoint,omitempty"`
		Size     int    `json:"size,omitempty"`
	} `json:"redisProxy,omitempty"`
	Proxy struct {
		Endpoint string `json:"endpoint,omitempty"`
		Size     int    `json:"size,omitempty"`
	} `json:"proxy,omitempty"`
}

func runList(ctx context.Context, o *clusteropts.ClusterOptions, typeFilter string, showPassword bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	var env envelope
	if err := client.DoJSON(ctx, "GET", "/middleware/v1/list", nil, &env); err != nil {
		return fmt.Errorf("list middleware: %w", err)
	}
	if env.Code != 0 && env.Code != 200 {
		// Server returned a structured failure inside a 2xx HTTP
		// envelope. Surface code + message so the user sees the
		// actionable error rather than an empty table.
		if env.Message != "" {
			return fmt.Errorf("list middleware: server returned code=%d: %s", env.Code, env.Message)
		}
		return fmt.Errorf("list middleware: server returned code=%d", env.Code)
	}

	items := env.Data
	if typeFilter != "" {
		filtered := make([]Item, 0, len(items))
		want := strings.ToLower(strings.TrimSpace(typeFilter))
		for _, it := range items {
			if strings.EqualFold(it.Type, want) {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}

	// Stable order across runs.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type
		}
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})

	if o.IsJSON() {
		out := items
		if !showPassword {
			out = redactPasswords(items)
		}
		return o.PrintJSON(out)
	}
	if o.Quiet {
		return nil
	}
	return renderListTable(items, o.NoHeaders)
}

func renderListTable(items []Item, noHeaders bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "TYPE\tNAME\tNAMESPACE\tNODES\tADMIN-USER")
	}
	for _, it := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			clusteropts.DashIfEmpty(it.Type),
			clusteropts.DashIfEmpty(it.Name),
			clusteropts.DashIfEmpty(it.Namespace),
			it.Nodes,
			clusteropts.DashIfEmpty(it.AdminUser),
		)
	}
	if len(items) == 0 {
		w.Flush()
		fmt.Fprintln(os.Stderr, "no middleware instances visible to this profile")
	}
	return nil
}

// redactPasswords returns a copy of items with Password replaced by
// the literal "<redacted>". We never mutate the caller's slice so
// re-rendering with --show-passwords from the same fetch (future
// repeated invocations of the run loop) stays safe.
func redactPasswords(items []Item) []Item {
	out := make([]Item, len(items))
	for i, it := range items {
		if it.Password != "" {
			it.Password = "<redacted>"
		}
		out[i] = it
	}
	return out
}
