package download

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// knownDownloadProviders is the set of download_provider values the
// server filters /api/download/unfinished on (models.DownloadProvider*
// in download-server). The endpoint requires exactly one provider per
// request and has no "all" mode, so when the caller omits --provider we
// fan out one request per provider and merge.
var knownDownloadProviders = []string{"yt-dlp", "aria2", "huggingface"}

func NewUnfinishedCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		provider string
		output   string
	)
	cmd := &cobra.Command{
		Use:   "unfinished",
		Short: "list tasks that have not reached a terminal state",
		Long: `List unfinished download tasks (GET /api/download/unfinished).

The server endpoint filters by a single required 'provider' (one of:
yt-dlp, aria2, huggingface) and does not accept an app filter. Pass
--provider to scope to one; omit it to query every provider and merge
the results. Only your own tasks are returned (scoped by the
gateway-injected identity).`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runUnfinished(c.Context(), f, provider, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&provider, "provider", "", "download provider to scope to (one of: yt-dlp, aria2, huggingface); omit for all")
	return cmd
}

// resolveUnfinishedProviders validates --provider and returns the list
// of providers to query: the single requested one, or all known
// providers when the flag is empty.
func resolveUnfinishedProviders(provider string) ([]string, error) {
	p := strings.TrimSpace(provider)
	if p == "" {
		return knownDownloadProviders, nil
	}
	for _, known := range knownDownloadProviders {
		if p == known {
			return []string{p}, nil
		}
	}
	return nil, fmt.Errorf("unknown --provider %q (allowed: %s)", provider, strings.Join(knownDownloadProviders, ", "))
}

func runUnfinished(ctx context.Context, f *cmdutil.Factory, provider, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	providers, err := resolveUnfinishedProviders(provider)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	// The endpoint returns {code, data: [tasks]} (a bare array under
	// "data"), not the {list,total} envelope the list endpoint uses,
	// so decode each provider's response straight into a slice.
	var all []DownloadTask
	for _, p := range providers {
		q := url.Values{}
		q.Set("provider", p)
		var tasks []DownloadTask
		if err := doGet(ctx, pc.doer, "/api/download/unfinished"+encodeQuery(q), &tasks); err != nil {
			return err
		}
		all = append(all, tasks...)
	}
	// Match the single-provider server ordering (updated_at ASC) across
	// the merged set so the combined view stays stable.
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].UpdatedAt.Before(all[j].UpdatedAt)
	})

	result := ListResult{List: all, Total: int64(len(all))}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderListTable(os.Stdout, result)
	}
}
